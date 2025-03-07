package main

// сюда писать код

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type Task struct {
	ID         int
	Сreator    string
	Title      string
	Performers string
	Doing      bool
}

type TaskManager struct {
	BotToken   string
	WebhookURL string
	taskMux    sync.Mutex
	tasks      []Task
	usersID    map[string]int64
}

var (
	// go run bot.go -tg.token="7558915533:AAHoZHhJTiEh6SPd26OS_W4ejo6tm9R4HQU" -tg.webhook="https://69c0-176-58-119-28.ngrok-free.app"
	// @BotFather в телеграме даст вам это
	BotToken = ""

	// урл выдаст вам игрок или хероку
	// напишите, когда будете проверять, я создам новый WebhookURL

	WebhookURL  = ""
	nothingTask = "Нет задач"
	correctID   = "Ввeдите корректно ID задачи"
)

// одна задача для OwnerTask
func generateAnswerOwnerTask(task Task) string {
	var strAns bytes.Buffer
	taskID := strconv.Itoa(task.ID)

	strAns.WriteString(taskID)
	strAns.WriteString(". ")
	strAns.WriteString(task.Title)
	strAns.WriteString(" by @")
	strAns.WriteString(task.Сreator)
	strAns.WriteString("\n")

	if task.Performers == "" {
		strAns.WriteString("/assign_")
		strAns.WriteString(taskID)

		return strAns.String()
	}

	strAns.WriteString("/unassign_")
	strAns.WriteString(taskID)
	strAns.WriteString(" /resolve_")
	strAns.WriteString(taskID)
	strAns.WriteString("\n")

	return strAns.String()
}

// выводит задачи созданные пользователем
func (tm *TaskManager) ownerTask(userName string) string {
	tm.taskMux.Lock()
	nowTask := tm.tasks
	tm.taskMux.Unlock()

	if len(tm.tasks) == 0 {
		return nothingTask
	}

	var strAns bytes.Buffer

	for _, task := range nowTask {
		if task.Doing || task.Сreator != userName {
			continue
		}
		oneStr := generateAnswerOwnerTask(task)
		strAns.WriteString(oneStr)
	}

	if len(strAns.String()) == 0 {
		return nothingTask
	}

	return strAns.String()
}

// одна задача для myTask
func generateAnswerMyTask(task Task) string {
	var strAns bytes.Buffer
	taskID := strconv.Itoa(task.ID)

	strAns.WriteString(taskID)
	strAns.WriteString(". ")
	strAns.WriteString(task.Title)
	strAns.WriteString(" by @")
	strAns.WriteString(task.Сreator)
	strAns.WriteString("\n")
	strAns.WriteString("/unassign_")
	strAns.WriteString(taskID)
	strAns.WriteString(" /resolve_")
	strAns.WriteString(taskID)
	strAns.WriteString("\n")
	strAns.WriteString("\n")

	return strAns.String()
}

// задачи где пользователь исполнитель
func (tm *TaskManager) myTask(userName string) string {
	tm.taskMux.Lock()
	nowTask := tm.tasks
	tm.taskMux.Unlock()

	if len(tm.tasks) == 0 {
		return nothingTask
	}

	var strAns bytes.Buffer

	for _, task := range nowTask {
		if task.Doing || task.Performers != userName {
			continue
		}
		oneStr := generateAnswerMyTask(task)
		strAns.WriteString(oneStr)
	}

	if len(strAns.String()) == 0 {
		return nothingTask
	}

	return strAns.String()[:len(strAns.String())-2]
}

// выполненение задачи
func (tm *TaskManager) resolveTask(idStr string, userName string) map[int64]string {
	ans := make(map[int64]string)
	ID, err := strconv.Atoi(idStr)
	if err != nil || ID < len(tm.tasks) {
		ans[tm.usersID[userName]] = correctID
		return ans
	}

	ID--

	if tm.tasks[ID].Performers != userName {
		ans[tm.usersID[userName]] = "Задача не на вас"
		return ans
	}

	if tm.tasks[ID].Doing {
		ans[tm.usersID[userName]] = "Задача уже была выполнена"
		return ans
	}

	tm.taskMux.Lock()
	task := tm.tasks[ID]
	tm.tasks[ID].Doing = true
	tm.taskMux.Unlock()

	var strAns bytes.Buffer

	strAns.WriteString("Задача \"")
	strAns.WriteString(task.Title)
	strAns.WriteString("\" выполнена")
	ans[tm.usersID[userName]] = strAns.String()

	if userName == task.Сreator {
		return ans
	}

	var strAns2 bytes.Buffer

	strAns2.WriteString("Задача \"")
	strAns2.WriteString(task.Title)
	strAns2.WriteString("\" выполнена @")
	strAns2.WriteString(userName)
	ans[tm.usersID[task.Сreator]] = strAns2.String()

	return ans
}

// снять исполнителя с задачи
func (tm *TaskManager) unassignTask(idStr string, userName string) map[int64]string {
	ans := make(map[int64]string)
	ID, err := strconv.Atoi(idStr)
	if err != nil || ID < len(tm.tasks) {
		ans[tm.usersID[userName]] = correctID
		return ans
	}

	ID--

	if tm.tasks[ID].Performers != userName {
		ans[tm.usersID[userName]] = "Задача не на вас"
		return ans
	}

	if tm.tasks[ID].Doing {
		ans[tm.usersID[userName]] = "Задача уже выполнена"
		return ans
	}

	tm.taskMux.Lock()
	task := tm.tasks[ID]
	tm.tasks[ID].Performers = ""
	tm.taskMux.Unlock()
	var strAns bytes.Buffer

	strAns.WriteString("Принято")
	ans[tm.usersID[userName]] = strAns.String()

	if userName == task.Сreator {
		return ans
	}

	var strAns2 bytes.Buffer
	strAns2.WriteString("Задача \"")
	strAns2.WriteString(task.Title)
	strAns2.WriteString("\" осталась без исполнителя")
	ans[tm.usersID[task.Сreator]] = strAns2.String()

	return ans
}

// стать исполнителем на задачу
func (tm *TaskManager) assignTask(idStr string, userName string) map[int64]string {
	ans := make(map[int64]string)
	ID, err := strconv.Atoi(idStr)
	if err != nil || ID > len(tm.tasks) || ID < 1 {
		ans[tm.usersID[userName]] = correctID
		return ans
	}

	ID--

	if tm.tasks[ID].Doing {
		ans[tm.usersID[userName]] = "Задача уже выполнена"
		return ans
	}

	tm.taskMux.Lock()
	task := tm.tasks[ID]
	tm.tasks[ID].Performers = userName
	tm.taskMux.Unlock()
	var strAns bytes.Buffer
	strAns.WriteString("Задача \"")
	strAns.WriteString(task.Title)
	strAns.WriteString("\" назначена на вас")
	ans[tm.usersID[userName]] = strAns.String()
	var strAns2 bytes.Buffer
	strAns2.WriteString("Задача \"")
	strAns2.WriteString(task.Title)
	strAns2.WriteString("\" назначена на @")
	strAns2.WriteString(userName)

	if task.Performers != "" && task.Performers != userName {
		ans[tm.usersID[task.Performers]] = strAns2.String()
		return ans
	} else if userName != tm.tasks[ID].Сreator {
		ans[tm.usersID[task.Сreator]] = strAns2.String()
	}

	return ans
}

// одна задача для allTasks
func generateAnswerAllTasks(task Task, user string) string {
	var strAns bytes.Buffer
	taskID := strconv.Itoa(task.ID)

	strAns.WriteString(taskID)
	strAns.WriteString(". ")
	strAns.WriteString(task.Title)
	strAns.WriteString(" by @")
	strAns.WriteString(task.Сreator)
	strAns.WriteString("\n")

	if task.Performers != "" {
		strAns.WriteString("assignee: ")

		if task.Performers == user {
			strAns.WriteString("я\n")
			strAns.WriteString("/unassign_")
			strAns.WriteString(taskID)
			strAns.WriteString(" /resolve_")
			strAns.WriteString(taskID)
		} else {
			strAns.WriteString("@")
			strAns.WriteString(task.Performers)
		}

	} else {
		strAns.WriteString("/assign_")
		strAns.WriteString(taskID)
	}
	strAns.WriteString("\n\n")

	return strAns.String()
}

// выводит все не выполненные задачи
func (tm *TaskManager) allTasks(user string) string {
	tm.taskMux.Lock()
	nowTask := tm.tasks
	tm.taskMux.Unlock()

	if len(tm.tasks) == 0 {
		return nothingTask
	}

	var strAns bytes.Buffer

	for _, task := range nowTask {
		if task.Doing {
			continue
		}
		oneStr := generateAnswerAllTasks(task, user)
		strAns.WriteString(oneStr)

	}

	if len(strAns.String()) == 0 {
		return nothingTask
	}

	return strAns.String()[:len(strAns.String())-2]
}

// создает новую задачу
func (tm *TaskManager) newTask(title string, user string) string {
	tm.taskMux.Lock()
	defer tm.taskMux.Unlock()
	n := len(tm.tasks)
	n++

	task := Task{
		ID:         n,
		Сreator:    user,
		Title:      title,
		Performers: "",
		Doing:      false,
	}

	tm.tasks = append(tm.tasks, task)
	var strAns bytes.Buffer

	strAns.WriteString("Задача \"")
	strAns.WriteString(task.Title)
	strAns.WriteString("\" создана, id=")
	strAns.WriteString(strconv.Itoa(task.ID))

	return strAns.String()
}

// обработка команд
func (tm *TaskManager) commandTask(command []string, str []string, chatID int64, userName string, bot *tgbotapi.BotAPI) error {
	var err error
	switch command[0] {
	case "/new":
		if len(str) == 1 || len(command) != 1 {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				"Ввeдите название после new через пробел",
			))
		} else {
			title := strings.Join(str[1:], " ")
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				tm.newTask(title, userName),
			))
		}

	case "/tasks":
		if len(command) != 1 {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				"Ввeдите просто команду /task",
			))
		} else {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				tm.allTasks(userName),
			))
		}

	case "/assign":
		if len(command) != 2 || len(str) != 1 {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				correctID,
			))
		} else {
			taskID := command[1]
			ans := tm.assignTask(taskID, userName)
			for key, value := range ans {
				_, err = bot.Send(tgbotapi.NewMessage(
					key,
					value,
				))
			}
		}

	case "/unassign":
		if len(command) != 2 || len(str) != 1 {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				correctID,
			))
		} else {
			taskID := command[1]
			ans := tm.unassignTask(taskID, userName)
			for key, value := range ans {
				_, err = bot.Send(tgbotapi.NewMessage(
					key,
					value,
				))
			}
		}

	case "/resolve":
		if len(command) != 2 || len(str) != 1 {
			_, err = bot.Send(tgbotapi.NewMessage(
				chatID,
				correctID,
			))
		} else {
			taskID := command[1]
			ans := tm.resolveTask(taskID, userName)
			for key, value := range ans {
				_, err = bot.Send(tgbotapi.NewMessage(
					key,
					value,
				))
			}
		}

	case "/my":
		_, err = bot.Send(tgbotapi.NewMessage(
			chatID,
			tm.myTask(userName),
		))

	case "/owner":
		_, err = bot.Send(tgbotapi.NewMessage(
			chatID,
			tm.ownerTask(userName),
		))

	default:
		_, err = bot.Send(tgbotapi.NewMessage(
			chatID,
			"введите корректную команду",
		))
	}

	return err
}

// создание бота
func startTaskBot(_ context.Context) error {
	BotTokenflag := flag.String("tg.token", "", "token for telegram")

	WebhookURLFlag := flag.String("tg.webhook", "", "webhook addr for telegram")
	flag.Parse()
	if BotToken == "" {
		BotToken = *BotTokenflag
	}

	if WebhookURL == "" {
		WebhookURL = *WebhookURLFlag
	}

	tm := TaskManager{
		BotToken:   BotToken,
		WebhookURL: WebhookURL,
		taskMux:    sync.Mutex{},
		tasks:      []Task{},
		usersID:    make(map[string]int64),
	}

	bot, err := tgbotapi.NewBotAPI(tm.BotToken)
	if err != nil {
		log.Fatalf("NewBotAPI failed: %s", err)
		return err
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(tm.WebhookURL)
	if err != nil {
		log.Fatalf("NewWebhook failed: %s", err)
		return err
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("SetWebhook failed: %s", err)
		return err
	}

	updates := bot.ListenForWebhook("/")

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte("all is working"))
		if err != nil {
			log.Fatalf("Write failed: %s", err)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)

	for update := range updates {
		messageText := update.Message.Text
		chatID := update.Message.Chat.ID
		userName := update.Message.From.UserName
		str := strings.Split(messageText, " ")
		command := strings.Split(str[0], "_")

		tm.usersID[userName] = chatID

		err := tm.commandTask(command, str, chatID, userName, bot)
		if err != nil {
			log.Printf("bot error %s", err)
			return err
		}

	}
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
