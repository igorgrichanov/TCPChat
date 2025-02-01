package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestOneClient(t *testing.T) {
	address := "localhost:8000"
	message1 := "hello from 1"
	in1 := strings.NewReader(message1)

	// создание и запуск сервера
	s := newServer(address)
	go s.start()

	// перенаправление вывода
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// подключение первого клиента
	cl1 := newClient(address, in1)
	go cl1.connectToChat()

	time.Sleep(time.Second)
	close(s.quitCh)
	time.Sleep(time.Second)
	w.Close()
	os.Stdout = old

	// проверка успешности теста
	correctOutput := "You are " + cl1.clientAddress + "\n" + cl1.clientAddress + ": " + message1 + "\n"

	var actualOutput bytes.Buffer
	if _, err := io.Copy(&actualOutput, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}

	if actualOutput.String() != correctOutput {
		t.Errorf("got %q, want %q", actualOutput.String(), correctOutput)
	}
}

func TestThreeClients(t *testing.T) {
	// ОС не сразу чистит файловый дескриптор, связанный с сокетом, поэтому выбираем другой порт
	address := "localhost:8001"

	// подготовка буферов
	message1 := "hello from 1"
	message2 := "hello from 2"
	message3 := "hello from 3"
	in1 := strings.NewReader(message1)
	in2 := strings.NewReader(message2)
	in3 := strings.NewReader(message3)

	// создание и запуск сервера
	s := newServer(address)
	go s.start()

	// подключение первого клиента
	cl1 := newClient(address, in1)
	go cl1.connectToChat()
	// спим для гарантии порядка подключения клиентов, чтобы протестить итоговый вывод программы
	time.Sleep(500 * time.Millisecond)

	// подключение второго клиента
	cl2 := newClient(address, in2)
	go cl2.connectToChat()
	time.Sleep(500 * time.Millisecond)

	// подключение третьего клиента
	cl3 := newClient(address, in3)
	go cl3.connectToChat()

	// отключение второго клиента
	time.Sleep(500 * time.Millisecond)
	close(cl2.quitCh)

	// отключение первого клиента
	time.Sleep(500 * time.Millisecond)
	close(cl1.quitCh)

	time.Sleep(500 * time.Millisecond)
	close(s.quitCh)
	time.Sleep(500 * time.Millisecond)
}
