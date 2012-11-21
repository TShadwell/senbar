package main
import "time"
import "sort"
type poller interface{
	poll()
	//Although this is a function, interval really should be constant
	interval() int
	String() string
}

type pollMachine struct{
	pollers map[int]*poller
	priorityKeys []int
}
type runningPollMachine struct{
	//Pollers in priority order
	pollers []*poller
	//A map of time (seconds) to pollers to poll
	pollOrder map[int][]*poller
	//Highest poll time
	highestPollTime int
}

func (pl *pollMachine) add(p *poller, priority int){ 
	if pl.pollers[priority] == nil{
		(*pl).pollers[priority] = p
		(*pl).priorityKeys = append((*pl).priorityKeys, priority)
	} else {
		panic("Cannot have two pollers with the same priority!")
	}
}
func (rpm *runningPollMachine) run(){
	for{
			for i:=0; i<= rpm.highestPollTime; i++{
				time.Sleep(1*time.Second)
				if rpm.pollOrder[i] != nil{
					for _, val := range rpm.pollOrder[i] {
						(*val).poll()
					}
				}
			}
	}
}
func (pl *pollMachine) start() (rpm *runningPollMachine, hasPollers bool){
	rpm = &runningPollMachine{
		make([]*poller, len(pl.priorityKeys)),
		make(map[int][]*poller),
		0,
	}
	hasPollers = false
	if len(pl.pollers) >0 {
		hasPollers = true
		sort.Ints(pl.priorityKeys)
		for i, val := range pl.priorityKeys{
			thisPoller := pl.pollers[val]
			interval := (*thisPoller).interval()
			//Add the pollers in priority order
			(*rpm).pollers[i] = thisPoller

			//Make the time map
			if rpm.pollOrder[interval]  == nil{

				(*rpm).pollOrder[interval] = make([]*poller, 1)

			}

			(*rpm).pollOrder[interval] = append(rpm.pollOrder[interval], thisPoller)
			//Record the highest time
			if rpm.highestPollTime < interval {
				(*rpm).highestPollTime = interval
			}

		}
		go (*rpm).run()
	}
	return
}

func (rpm *runningPollMachine) draw() (output string){
	output=""
	for _, p := range rpm.pollers {
		output += (*p).String()
	}
	return
}

var Pollers pollMachine = pollMachine{
	make(map[int]*poller),
	make([]int,0),
}




