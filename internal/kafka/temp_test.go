package kafka

import (
	"log"
	"math/rand"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	lastConsume := time.Now()
	interval := time.Second * 3
	for i := 0; i < 10; i++ {
		timer := time.NewTimer(lastConsume.Add(interval).Sub(time.Now()))
		<-timer.C
		lastConsume = time.Now()
		log.Println("Process...")
		sleepMills := time.Duration(rand.Intn(10000)) * time.Millisecond
		log.Println("Sleep", sleepMills)
		time.Sleep(sleepMills)
	}
}
