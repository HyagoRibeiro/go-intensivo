package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/HyagoRibeiro/go-intensivo/internal/infra/database"
	"github.com/HyagoRibeiro/go-intensivo/internal/usecase"
	"github.com/HyagoRibeiro/go-intensivo/pkg/rabbitmq"

	_ "github.com/mattn/go-sqlite3"

	amqp "github.com/rabbitmq/amqp091-go"
)


type Car struct {
	Model string
	Color string
}

// metodo

func (c Car) Start() {
	println(c.Model + "is started")
}

func (c *Car) ChangeColor(color string) {
	c.Color = color
	println("new color: " + c.Color)
}

//função

func soma(x, y int) int {
	return x + y
}

func main() {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	orderRepository := database.NewOrderRepository(db)
	uc := usecase.NewCalculateFinalPrice(orderRepository)
	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()
	msgRabbitmqChannel := make(chan amqp.Delivery)
	go rabbitmq.Consume(ch, msgRabbitmqChannel) //fica escutando a fila === trava. com Go abre uma T2
	rabbitmqWorker(msgRabbitmqChannel, uc) //T1

}

func rabbitmqWorker(msgChan chan amqp.Delivery, uc *usecase.CalculateFinalPrice) {
	fmt.Println("starting rabbitmq")
	for msg := range msgChan {
		var input usecase.OrderInput
		err := json.Unmarshal(msg.Body, &input)
		if err != nil {
			panic(err)
		}
		output, err := uc.Execute(input)
		if err != nil {
			panic(err)
		}
		msg.Ack(false)
		fmt.Println("Mensagem processada e salva no banco:", output)
	}
}
