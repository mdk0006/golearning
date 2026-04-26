package main

import "fmt"

func main() {
	onCallEngineers := []string{"alice", "bob", "carol"}
	fmt.Println("length of oncallEngineers slice", len(onCallEngineers))
	fmt.Println("cap of oncallEngineers slice", cap(onCallEngineers))
	onCallEngineers = append(onCallEngineers, "danish")
	fmt.Println("updated length of oncallEngineers slice", len(onCallEngineers))
	fmt.Println("updated cap of oncallEngineers slice", cap(onCallEngineers))
	firstTwoEng := onCallEngineers[0:2]
	fmt.Println("First two engineers in rotation", firstTwoEng)
	safe := make([]string, len(onCallEngineers))
	copy(safe, onCallEngineers)
	newOnCallSchedule := append(safe[:1], safe[2:]...)
	fmt.Println("firstTwoEng after remove:", firstTwoEng)
	fmt.Println("New Roster", newOnCallSchedule)

}
