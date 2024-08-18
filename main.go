package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// Used to store questions and answers from csv
type problem struct {
	question string
	answer   string
	counter
	scanner
}

type counter struct {
	count int64
}

type scanner struct {
	scan *bufio.Reader
}

func main() {
	var filenameStr = "filename"
	var timeLimitStr = "timeLimit"
	var defaultFilename = "problems.csv"
	var shuffle = "shuffle"
	var done = make(chan bool, 1)
	var problem = problem{}

	// Initialize flag
	filename := flag.String(filenameStr,
		defaultFilename,
		"File containing the questions and answers")

	timeLimit := flag.Int(timeLimitStr, 30, "time limit of the quiz")
	isShuffle := flag.Bool(shuffle, false, "Shuffle the question set")
	flag.Parse()

	input, err := problem.start()

	if err != nil {
		log.Fatal(err)
	}

	if len(*input) > 0 {
		timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
		// Read data from retrieved file
		data, err := readFile(*filename)

		if err != nil {
			log.Fatal(err)
		}

		if *isShuffle {

			data = shuffleData(*data)
		}

		// Iterate thru data and ask question
		go problem.askQuestions(data, done)

		defer close(done)

		select {
		case <-done:
			fmt.Println("Congrats")
		case <-timer.C:
			// Double check if done is received
			// in case of timeouts in test
			select {
			default:
			case <-done:
				fmt.Println("Congrats")
			}

			fmt.Println("timeout!")

			fmt.Printf("Sorry you ran out of time. Your score is %v/%v \n", problem.count, len(*data))
		}

	}

}

func shuffleData(d [][]string) *[][]string {
	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})

	return &d
}

// prompt to start the quiz
func (p *problem) start() (*string, error) {
	p.scan = bufio.NewReader(os.Stdin)

	fmt.Println("If ready, press any key to start.....")
	input, err := p.scan.ReadString('\n')

	if err != nil {
		return nil, err
	}

	return &input, nil

}

func (p *problem) askQuestion(q string) (*string, error) {
	p.scan = bufio.NewReader(os.Stdin)
	fmt.Println("Question: \n", q)
	input, err := p.scan.ReadString('\n')

	if err != nil {
		return nil, err
	}

	input = strings.ToLower(strings.TrimSpace(strings.TrimRight(input, "\n")))

	return &input, nil
}

func readFile(f string) (*[][]string, error) {
	file, err := os.Open(f)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, err := csv.NewReader(file).ReadAll()

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (p *problem) askQuestions(data *[][]string, done chan bool) {

	for _, value := range *data {

		p.question = value[0]
		p.answer = value[1]

		question, err := p.askQuestion(p.question)

		if err != nil {
			log.Fatal(err)
		}

		if *question == p.answer {
			fmt.Println("Correct!")
			p.counterAdd()
		} else {
			fmt.Println("Incorrect!")
		}

	}

	fmt.Printf("Quiz is done! Your score is %v/%v ", p.getCounter(), len(*data))

	done <- true

}

func (p *problem) counterAdd() {
	atomic.AddInt64(&p.count, 1)
}

func (p *problem) getCounter() int64 {
	return p.count
}
