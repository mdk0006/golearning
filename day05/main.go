// day05/main.go
package main

import "fmt"

type Rotation struct {
	Engineer string
	Start    string
	End      string
}

type Schedule struct {
	Name      string
	Rotations []Rotation
}

// TODO: write two versions
func (s Schedule) AddRotation(r Rotation) {
	s.Rotations = append(s.Rotations, r)
}

func (s *Schedule) AddRotationPtr(r Rotation) {
	s.Rotations = append(s.Rotations, r)
}

func main() {
	sched := Schedule{
		Name:      "Primary Oncall",
		Rotations: make([]Rotation, 0, 4),
	}
	fmt.Println("before - len:", len(sched.Rotations), "cap:", cap(sched.Rotations))
	sched.AddRotation(Rotation{Engineer: "Danish"})
	fmt.Println("after - len:", len(sched.Rotations), "cap", cap(sched.Rotations))
	fmt.Println("rotations:", sched.Rotations)
	fmt.Println("ghost:", sched.Rotations[:1])

}
