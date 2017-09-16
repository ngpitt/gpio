// MIT License

// Copyright (c) 2017 Alex Ellis
// Copyright (c) 2017 Isaac "Ike" Arias
// Copyright (c) 2017 Nicholas Pitt

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gpio

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type GpioObj struct {
	pinDescriptors map[int]pinDescriptor
}

type Gpio interface {
	Write(pin int, value int)
	Cleanup()
}

type pinDescriptor struct {
	value     *os.File
	direction *os.File
}

func NewGpio() Gpio {
	return &GpioObj{make(map[int]pinDescriptor)}
}

func (o *GpioObj) Write(pin int, value int) {
	pd, exists := o.pinDescriptors[pin]
	if !exists {
		if !isExported(pin) {
			export(pin)
		}
		path := fmt.Sprintf("/sys/class/gpio/gpio%d", pin)
		value, err := os.OpenFile(path+"/value", os.O_WRONLY, 0640)
		if err != nil {
			log.Panicln(err.Error())
		}
		direction, err := os.OpenFile(path+"/direction", os.O_WRONLY, 0640)
		if err != nil {
			log.Panicln(err.Error())
		}
		_, err = direction.Write([]byte("out"))
		if err != nil {
			log.Panicln(err.Error())
		}
		pd = pinDescriptor{value, direction}
		o.pinDescriptors[pin] = pd
	}
	_, err := pd.value.Write([]byte(strconv.Itoa(value)))
	if err != nil {
		log.Panicln(err.Error())
	}
}

func (o *GpioObj) Cleanup() {
	for p, pd := range o.pinDescriptors {
		err := pd.direction.Close()
		if err != nil {
			log.Panicln(err.Error())
		}
		err = pd.value.Close()
		if err != nil {
			log.Panicln(err.Error())
		}
		unexport(p)
	}
}

func export(pin int) {
	err := ioutil.WriteFile("/sys/class/gpio/export", []byte(strconv.Itoa(pin)), 0644)
	if err != nil {
		log.Panicln(err.Error())
	}
}

func unexport(pin int) {
	err := ioutil.WriteFile("/sys/class/gpio/unexport", []byte(strconv.Itoa(pin)), 0644)
	if err != nil {
		log.Panicln(err.Error())
	}
}

func isExported(pin int) bool {
	_, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%d", pin))
	return !os.IsNotExist(err)
}
