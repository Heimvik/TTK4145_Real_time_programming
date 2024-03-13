package elevator

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4

var _mtx sync.Mutex
var _conn net.Conn

func F_InitDriver(addr string) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func F_SetMotorDirection(dir T_ElevatorDirection) {
	f_write([4]byte{1, byte(dir), 0, 0})
}

func F_SetButtonLamp(button T_ButtonType, floor int, value bool) {
	f_write([4]byte{2, byte(button), byte(floor), f_toByte(value)})
}

func F_SetFloorIndicator(floor int) {
	f_write([4]byte{3, byte(floor), 0, 0})
}

func F_SetDoorOpenLamp(value bool) {
	f_write([4]byte{4, f_toByte(value), 0, 0})
}

func F_SetStopLamp(value bool) {
	f_write([4]byte{5, f_toByte(value), 0, 0})
}

func F_PollButtons(receiver chan<- T_ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := T_ButtonType(0); b < 3; b++ {
				v := f_GetButton(b, f)
				if v != prev[f][b] && v {
					receiver <- T_ButtonEvent{f, T_ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func F_PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := f_GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func F_PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := f_GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func F_PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := f_GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func f_GetButton(button T_ButtonType, floor int) bool {
	a := f_read([4]byte{6, byte(button), byte(floor), 0})
	return f_toBool(a[1])
}

func f_GetFloor() int {
	a := f_read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func f_GetStop() bool {
	a := f_read([4]byte{8, 0, 0, 0})
	return f_toBool(a[1])
}

func f_GetObstruction() bool {
	a := f_read([4]byte{9, 0, 0, 0})
	return f_toBool(a[1])
}

func f_read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func f_write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func f_toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func f_toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
