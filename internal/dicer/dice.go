package dicer

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var regex = regexp.MustCompile(
	`(?P<num>\d+)d(?P<sides>\d+)(?P<keep>k\d{1,3})?((?P<op>[\+-])(?P<mod>\d+))?`)

const (
	Add      = "+"
	Subtract = "-"
)

type Emoji struct {
	Total   string
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

func New(text string) []*Dice {
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
		r.Emoji.Results = append(r.Emoji.Results, emoji(i))
	}
	for _, i := range r.Removed {
		r.Emoji.Removed = append(r.Emoji.Removed, emoji(i))
	}

	r.Emoji.Total = emoji(r.Total)

	return r
}

func (r *Dice) String() string {
	res := fmt.Sprintf("%dd%d", r.Number, r.Sides)

	if r.Keep > 0 {
		res += fmt.Sprintf("k%d", r.Keep)
	}

	if r.Mod != 0 {
		res += fmt.Sprintf("%s%d", r.Operator, r.Mod)
	}

	return res
}

func (r *Dice) HTML() string {
	die := r.String()
	die = strings.ReplaceAll(die, `+`, `\+`)
	die = strings.ReplaceAll(die, `-`, `\-`)
	return fmt.Sprintf(
		"<i>%s</i>\n <b>total</b>:<code>%d</code>, <b>rolls</b>:<code>%s</code>\n",
		die,
		r.Total,
		joinInts(r.Results, ", "),
	)
}

func (r *Dice) Markdown() string {
	die := r.String()
	die = strings.ReplaceAll(die, `+`, `\+`)
	die = strings.ReplaceAll(die, `-`, `\-`)
	return fmt.Sprintf(
		"%s\n *total*:%d, *rolls*:%s\n",
		die,
		r.Total,
		joinInts(r.Results, ", "),
	)
}

func (r *Dice) MarkdownEmoji() string {
	return fmt.Sprintf(
		"%s\n *total*:%s, *rolls*:%s\n",
		r.String(),
		r.Emoji.Total,
		strings.Join(r.Emoji.Results, ", "),
	)
}

func joinInts(nums []int, sep string) string {
	res := ""
	for l, i := len(nums)-1, 0; i <= l; i++ {
		if l == i {
			res += fmt.Sprintf("%d", nums[i])
		} else {
			res += fmt.Sprintf("%d, ", nums[i])
		}
	}
	return res
}

func emoji(i int) string {
	return fmt.Sprintf(`%d\u20E3`, i)
}
