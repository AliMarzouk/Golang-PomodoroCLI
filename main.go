package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"golang.org/x/term"
)

func playSoundNotification() {
	f, err := os.Open("alarm.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

type KeyboardInput int

const (
	up KeyboardInput = iota
	down
	kill
	enter
)

func displayCountDownWithMenu(countDownValue time.Duration, title string, isPaused bool, highlightedOptionP *int) {
	*highlightedOptionP = (*highlightedOptionP + 3) % 3
	var message string
	var options []string
	if isPaused {
		message = "Timer paused !"
		options = []string{"Stop", "Resume", "Quit"}
	} else {
		message = "Timer running ..."
		options = []string{"Stop", "Pause", "Quit"}
	}
	fmt.Print("\033[H\033[2J")
	fmt.Printf("[%6s] %v : %v \r\n", countDownValue, title, message)
	for index, option := range options {
		if index == *highlightedOptionP {
			fmt.Print(">>>")
		} else {
			fmt.Print("   ")
		}
		fmt.Print(option + "\r\n")
	}
}

func startCountDown(durationInMinutes int, title string, keyInputChannel chan KeyboardInput) bool {
	ticker := time.NewTicker(1 * time.Second)

	countDownDuration := time.Duration(durationInMinutes) * time.Minute
	start := time.Now()
	calculateRemainingTime := func() time.Duration {
		return time.Until(start.Add(countDownDuration)).Round(time.Second)
	}
	remainingTime := calculateRemainingTime()
	isPaused := false

	highlightedOption := 0

	displayCountDownWithMenu(remainingTime, title, isPaused, &highlightedOption)

	for remainingTime > 0 {
		select {
		case keyPressed := <-keyInputChannel:
			switch keyPressed {
			case up:
				highlightedOption -= 1
			case down:
				highlightedOption += 1
			case enter:
				switch highlightedOption {
				case 0:
					return false
				case 1:
					if !isPaused {
						isPaused = true
					} else {
						isPaused = false
						start = time.Now()
						countDownDuration = remainingTime
					}
				case 2:
					return true
				}
			case kill:
				return true
			}
			displayCountDownWithMenu(remainingTime, title, isPaused, &highlightedOption)
		case <-ticker.C:
			if !isPaused {
				remainingTime = calculateRemainingTime()
				displayCountDownWithMenu(remainingTime, title, false, &highlightedOption)
			}
		}
	}
	go playSoundNotification()

	ticker.Stop()
	return false
}

func readSingleCharacter(keyBoardInputChannel chan KeyboardInput) {
	for {
		b := make([]byte, 3)
		_, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		if b[0] == '\x1b' && b[1] == '[' && b[2] == 'A' {
			keyBoardInputChannel <- up
		}
		if b[0] == '\x1b' && b[1] == '[' && b[2] == 'B' {
			keyBoardInputChannel <- down
		}
		if b[0] == '\x03' {
			keyBoardInputChannel <- kill
		}
		if b[0] == '\r' {
			keyBoardInputChannel <- enter
		}
	}
}

func printMainMenu(highlightedOption int) {
	options := []string{"Focus time (25 min)", "Long break (15 min)", "Small break (5 min)"}

	fmt.Print("\033[H\033[2J")
	for index, option := range options {
		if index == (highlightedOption+len(options))%len(options) {
			fmt.Print(">>>")
		} else {
			fmt.Print("   ")
		}
		fmt.Print(option + "\r\n")
	}
}

func mainMenu(keyBoardInputChannel chan KeyboardInput) {
	highlightedMainOption := 0
	defer fmt.Print("Thank you for using the application\r\nMore infos at https://github.com/AliMarzouk\r\n")
	killed := false

	for !killed {
		printMainMenu(highlightedMainOption)
		keyPressed := <-keyBoardInputChannel
		switch keyPressed {
		case up:
			highlightedMainOption -= 1
		case down:
			highlightedMainOption += 1
		case enter:
			switch highlightedMainOption {
			case 0:
				killed = startCountDown(25, "Focus time", keyBoardInputChannel)
			case 1:
				killed = startCountDown(15, "Long break", keyBoardInputChannel)
			case 2:
				killed = startCountDown(1, "Small break", keyBoardInputChannel)
			}
		case kill:
			return
		}
	}
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	if err != nil {
		panic(err)
	}

	keyBoardInputChannel := make(chan KeyboardInput)
	go readSingleCharacter(keyBoardInputChannel)

	mainMenu(keyBoardInputChannel)
}
