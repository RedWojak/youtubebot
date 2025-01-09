package main

import (
	"fmt"
	"os"
	"youtubedownloader/telegrambot"
)

func main() {
	fmt.Println("hello")
	makeDirectoryIfNotExists("audio")
	makeDirectoryIfNotExists("video")
	makeDirectoryIfNotExists("output")
	
	bot := telegrambot.Telegrambot{}
	bot.NewBot()

	

	
}





func makeDirectoryIfNotExists(path string) error {
	
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}