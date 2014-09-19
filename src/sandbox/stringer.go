package main

import "fmt"

type Stringer interface {
	String() string
}

type Complex struct {
	first int
}

func (c Complex) String() string {
	return fmt.Sprintf("I implement h͉̠͇̹̖̼̖͈̬̘͕͎͈̬̺̺̯̅̃ͯ̈̂̐̿̍̌͑ͩ̑ͦ̿͡e̛̛̹̞͚̭̻͇̰̟͙̱͉̣̼̻̝͕̭ͯͦ͌̒̽̐l̸̵̷̛͍̬͎̲̲̹̙͙̞͚̲̭̫̬ͭ͗̒ͯ̈́ͣ͂ͧ̌̿ͮ̈́ͫ̽ͩ̈͞ṗ̨̟̰͈̹̗͖̰̙͉͎̗͓̬̳ͤ͒ͣ̓̂͛͑̑ͦͮ̅͆̈ͥ̈́́̈̋́̕͞ : %d", c.first)
}

func main() {
	container := Complex{first: 256}
	fmt.Println(container)
}
