package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// структура для всех игровых комнат
type Room struct {
	name      string
	doors     []string
	onTable   []string
	onChair   []string
	open      bool
	textForGo string
	inRoom    []string
}

// структура для самого игрока
type Player struct {
	inBackpack []string
	placeNow   *Room
	backpack   bool
}

// глобальные переменные для игрока и всех комнат
var player = new(Player)
var kitchen = new(Room)
var hallway = new(Room)
var myRoom = new(Room)
var street = new(Room)

// map для перемещения между комнатами
var doors map[string]*Room = map[string]*Room{
	"кухня":   kitchen,
	"комната": myRoom,
	"коридор": hallway,
	"улица":   street,
	"дом":     hallway,
	"дверь":   street,
}

// вспомогательная функция для удаляения элемента из слайса
func remove(sl []string, n int) []string {
	ans := []string{}
	if n > 0 {
		ans = append(ans, sl[0:n]...)
	}
	if n < len(sl)-1 {
		ans = append(ans, sl[n+1:]...)
	}
	return ans
}

// функция идти
func (player *Player) doGo(room string) string {
	var strAns bytes.Buffer
	place := *player.placeNow
	for _, v := range place.doors {
		if v == room {
			checkingRoom := *doors[room]
			if !checkingRoom.open {
				return "дверь закрыта"
			} else {
				player.placeNow = doors[room]
				strAns.WriteString(player.placeNow.textForGo)
				strAns.WriteString("можно пройти - ")
				for _, val := range checkingRoom.doors {
					strAns.WriteString(val)
					strAns.WriteString(", ")
				}
				return strAns.String()[:len(strAns.String())-2]
			}
		}
	}
	strAns.WriteString("нет пути в ")
	strAns.WriteString(room)
	return strAns.String()
}

// функция осмотреться
func (player Player) look() string {
	kitchenName := "Кухня"
	var ans bytes.Buffer
	place := player.placeNow
	lenTable := len(place.onTable)
	lenChair := len(place.onChair)
	if place.name == kitchenName {
		ans.WriteString("ты находишься на кухне, ")
	}
	if lenTable > 0 {
		ans.WriteString("на столе: ")
		for _, v := range place.onTable[:lenTable-1] {
			ans.WriteString(v)
			ans.WriteString(", ")
		}
		ans.WriteString(place.onTable[lenTable-1])
	}
	if lenChair > 0 {
		if len(ans.String()) > 0 {
			ans.WriteString(", ")
		}
		ans.WriteString("на стуле: ")
		for _, v := range place.onChair[:lenChair-1] {
			ans.WriteString(v)
			ans.WriteString(", ")
		}
		ans.WriteString(place.onChair[lenChair-1])
	}

	if lenTable+lenChair == 0 {
		ans.WriteString("пустая комната")
	}
	if place.name == kitchenName {
		if len(ans.String()) > 0 {
			ans.WriteString(", ")
		}
		ans.WriteString("надо ")
		check := false
		for _, v := range player.inBackpack {
			if v == "конспекты" {
				check = true
			}
		}
		if !check {
			ans.WriteString("собрать рюкзак и ")
		}
		ans.WriteString("идти в универ")
	}
	if len(ans.String()) > 0 {
		ans.WriteString(". ")
	}
	ans.WriteString("можно пройти - ")
	for _, val := range place.doors {
		ans.WriteString(val)
		ans.WriteString(", ")
	}
	return ans.String()[:len(ans.String())-2]

}

// функция надеть
func (player *Player) dress(subject string) string {
	var ans bytes.Buffer
	place := *player.placeNow
	for i, v := range place.onChair {
		if subject == v {
			player.inBackpack = append(player.inBackpack, subject)
			player.placeNow.onChair = remove(player.placeNow.onChair, i)
			if subject == "рюкзак" {
				player.backpack = true
			}
			ans.WriteString("вы надели: ")
			ans.WriteString(subject)
			return ans.String()
		}
	}
	for i, v := range place.onTable {
		if subject == v {
			player.inBackpack = append(player.inBackpack, subject)
			player.placeNow.onTable = remove(player.placeNow.onTable, i)
			if subject == "рюкзак" {
				player.backpack = true
			}
			ans.WriteString("вы надели: ")
			ans.WriteString(subject)
			return ans.String()
		}
	}
	return "нет такого"
}

// функция взять
func (player *Player) take(subject string) string {
	var ans bytes.Buffer
	place := *player.placeNow
	for i, v := range place.onChair {
		if subject == v {
			if !player.backpack {
				return "некуда класть"
			}
			player.inBackpack = append(player.inBackpack, subject)
			player.placeNow.onChair = remove(player.placeNow.onChair, i)
			ans.WriteString("предмет добавлен в инвентарь: ")
			ans.WriteString(subject)
			return ans.String()
		}
	}
	for i, v := range place.onTable {
		if subject == v {
			if !player.backpack {
				return "некуда класть"
			}
			player.inBackpack = append(player.inBackpack, subject)
			player.placeNow.onTable = remove(player.placeNow.onTable, i)
			ans.WriteString("предмет добавлен в инвентарь: ")
			ans.WriteString(subject)
			return ans.String()
		}
	}
	return "нет такого"
}

// функция применить
func (player Player) use(subject, appointment string) string {
	var ans bytes.Buffer
	availability := false
	for _, v := range player.inBackpack {
		if v == subject {
			availability = true
			break
		}
	}
	if !availability {
		ans.WriteString("нет предмета в инвентаре - ")
		ans.WriteString(subject)
		return ans.String()
	}
	for _, v := range player.placeNow.inRoom {
		if v == appointment {
			if appointment == "дверь" {
				doors[appointment].open = true
			}
			ans.WriteString(appointment)
			ans.WriteString(" открыта")
			return ans.String()
		}
	}
	return "не к чему применить"
}

/*
   код писать в этом файле
   наверняка у вас будут какие-то структуры с методами, глобальные переменные ( тут можно ), функции
*/

// наинаем игру с помощью initGame и считываем команды пользователя
func main() {
	initGame()

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		txt := in.Text()
		fmt.Println(handleCommand(txt))

	}

	/*
	   в этой функции можно ничего не писать,
	   но тогда у вас не будет работать через go run main.go
	   очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
}

// задаем начальные значения
func initGame() {

	/*
	   эта функция инициализирует игровой мир - все комнаты
	   если что-то было - оно корректно перезатирается
	*/

	player.inBackpack = []string{}
	player.placeNow = kitchen
	player.backpack = false

	kitchen.name = "Кухня"
	kitchen.doors = []string{"коридор"}
	kitchen.onTable = []string{"чай"}
	kitchen.onChair = []string{}
	kitchen.open = true
	kitchen.textForGo = "кухня, ничего интересного. "
	kitchen.inRoom = []string{}

	hallway.name = "Коридор"
	hallway.doors = []string{"кухня", "комната", "улица"}
	hallway.onTable = []string{}
	hallway.onChair = []string{}
	hallway.open = true
	hallway.textForGo = "ничего интересного. "
	hallway.inRoom = []string{"дверь"}

	myRoom.name = "Комната"
	myRoom.doors = []string{"коридор"}
	myRoom.onTable = []string{"ключи", "конспекты"}
	myRoom.onChair = []string{"рюкзак"}
	myRoom.open = true
	myRoom.textForGo = "ты в своей комнате. "
	myRoom.inRoom = []string{}

	street.name = "Улица"
	street.doors = []string{"домой"}
	street.onTable = []string{}
	street.onChair = []string{}
	street.open = false
	street.textForGo = "на улице весна. "
	street.inRoom = []string{}
}

// обработка команд от пользователя и вызов соответсвующих функций
func handleCommand(command string) string {

	/*
	   данная функция принимает команду от "пользователя"
	   и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/

	notValid := "неизвестная команда"

	res := strings.Split(command, " ")
	switch res[0] {
	case "идти":
		if len(res) > 1 {
			return player.doGo(res[1])
		} else {
			return notValid
		}
	case "надеть":
		if len(res) > 1 {
			return player.dress(res[1])
		} else {
			return notValid
		}
	case "взять":
		if len(res) > 1 {
			return player.take(res[1])
		} else {
			return notValid
		}
	case "осмотреться":
		return player.look()
	case "применить":
		if len(res) > 2 {
			return player.use(res[1], res[2])
		} else {
			return notValid
		}
	default:
		return notValid

	}
}
