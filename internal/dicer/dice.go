package dicer

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
)

var regex = regexp.MustCompile(
	`(?P<num>\d+)d(?P<sides>\d+)(?P<keep>k\d{1,3})?((?P<op>[\+-])(?P<mod>\d+))?`)

const (
	Add      = "+"
	Subtract = "-"
)

type Emoji struct {
	Results []string
	Removed []string
}

type Dice struct {
	Number   int
	Sides    int
	Keep     int
	Total    int
	Mod      int
	Results  []int
	Removed  []int
	Emoji    Emoji
	Operator string
}

func Parse(text string) []*Dice {
	var rolls []*Dice
	for _, m := range regex.FindAllStringSubmatch(text, 8) {
		dice := &Dice{
			Operator: Add,
			Number:   2,
			Sides:    6,
		}
		for i, name := range regex.SubexpNames() {
			switch name {
			case "op":
				if m[i] == Add {
					dice.Operator = Add
				} else if m[i] == Subtract {
					dice.Operator = Subtract
				}
			case "mod":
				num, err := strconv.Atoi(m[i])
				if err != nil {
					num = 0
				}
				dice.Mod = num
			case "num":
				num, err := strconv.Atoi(m[i])
				if err != nil {
					num = 1
				}
				if num > 100 {
					num = 100
				}
				dice.Number = num
			case "sides":
				if m[i] == "" {
					dice.Sides = 6
				} else {
					dice.Sides, _ = strconv.Atoi(m[i])
					if dice.Sides < 1 {
						dice.Sides = 0
					}
					if dice.Sides > 1000000 {
						dice.Sides = 1000000
					}
				}
			case "keep":
				if m[i] == "" {
					break
				}
				dice.Keep, _ = strconv.Atoi(m[i][1:])
				if dice.Keep > dice.Number {
					dice.Keep = dice.Number
				} else if dice.Keep < -dice.Number {
					dice.Keep = -dice.Number
				}
			}
		}

		// perform calculations please!
		dice.roll()
		rolls = append(rolls, dice)
	}

	return rolls
}

func (r *Dice) roll() *Dice {
	r.Results = []int{}

	if r.Sides == 0 {
		r.Results = append(r.Results, 0)
	} else if r.Sides == 1 {
		r.Results = append(r.Results, r.Number)
	} else {
		num := r.Number
		for i := 0; i < num; i++ {
			n := rand.Intn(r.Sides) + 1
			r.Results = append(r.Results, n)
		}
	}

	if r.Keep != 0 {
		sort.Ints(r.Results)

		if r.Keep > 0 {
			split := len(r.Results) - r.Keep
			r.Removed = r.Results[:split]
			r.Results = r.Results[split:]
		} else {
			split := -r.Keep
			r.Removed = r.Results[split:]
			r.Results = r.Results[:split]
		}
	}

	r.Total = 0
	if r.Operator == Add {
		r.Total += r.Mod
	} else {
		r.Total -= r.Mod
	}
	for _, i := range r.Results {
		r.Total += i
	}

	for _, i := range r.Results {
		r.Emoji.Results = append(r.Emoji.Results, fmt.Sprintf(`%d\u20E3`, i))
	}
	for _, i := range r.Removed {
		r.Emoji.Removed = append(r.Emoji.Removed, fmt.Sprintf(`%d\u20E3`, i))
	}

	return r
}

func (r *Dice) AsEmoji() []string {

	return emojis
}
