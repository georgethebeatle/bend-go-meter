package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Note struct {
	Freq  float64
	Floor float64
	Ceil  float64
	Names []string
}

// encode some note values
var notes []Note = make([]Note, 3)
var noteNames map[string]int = make(map[string]int)

func initNotes() {
	// TODO: automate this painstaking process by looping over a file
	// containing records {[notenames], freq} - pre-calculate Floor and Ceil
	// once on program's startup
	noteNames["C4"] = 0
	noteNames["Db4"] = 1
	noteNames["D4"] = 2

	notes[0] = Note{Freq: 261.626, Floor: 254.284, Ceil: 269.405, Names: []string{"C4"}}
	notes[1] = Note{Freq: 277.183, Floor: 269.405, Ceil: 285.424, Names: []string{"Db4"}}
	notes[2] = Note{Freq: 293.665, Floor: 285.424, Ceil: 302.396, Names: []string{"D4"}}
}

func main() {
	// sox will listen to built in mic and write contents to a pipe
	soxCmd := exec.Command("sox", "-q", "-d", "-t", "wav", "-")
	micIn, err := soxCmd.StdoutPipe()
	failOn(err)
	defer micIn.Close()

	// aubiopitch will read what sox streams in and determine pitch
	pitchCmd := exec.Command("aubiopitch", "-i", "-")
	pitchCmd.Stdin = micIn
	out, err := pitchCmd.StdoutPipe()
	failOn(err)
	defer out.Close()

	must(soxCmd.Start())
	go soxCmd.Wait()

	must(pitchCmd.Start())
	go pitchCmd.Wait()

	pitchOut := bufio.NewReader(out)

	initNotes()
	for {
		lineBytes, _, err := pitchOut.ReadLine()
		failOn(err)
		line := string(lineBytes)

		lineVec := strings.Split(line, " ")
		if len(lineVec) != 2 {
			continue
		}
		noteFreq, err := strconv.ParseFloat(lineVec[1], 64)
		if err != nil {
			continue
		}

		noteNames, dev := findNote(noteFreq)
		fmt.Printf("Names: %#v, Deviation: %f\n", noteNames, dev)
	}
}

func findNote(freq float64) ([]string, float64) {
	matchingNoteIdx := -1
	for i, candidateNote := range notes {
		if freq > candidateNote.Floor && freq < candidateNote.Ceil {
			matchingNoteIdx = i
		}
	}

	if matchingNoteIdx < 0 {
		return []string{}, 0.0
	}

	matchingNote := notes[matchingNoteIdx]
	var dev float64 = freq - matchingNote.Freq
	if dev > 0 {
		dev = dev / (matchingNote.Ceil - matchingNote.Freq)
	} else {
		dev = dev / (matchingNote.Freq - matchingNote.Floor)
	}

	return matchingNote.Names, dev
}

func must(err error) {
	if err != nil {
		fmt.Println("ERR", err.Error())
		panic(err.Error())
	}
}

var failOn func(err error) = must
