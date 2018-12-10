package spool

import (
	"testing"
	"io"
	"fmt"
	"strconv"
	"sync"
	"math/rand"
)

var count int

func TestNew(t *testing.T) {

	i := New(func() (io.Closer, error) {
		id := rand.Intn(50)
		return db{id}, nil
	}, 3)

	wg := sync.WaitGroup{}
	wg.Add(10)

	//开启5个协程跑
	for e := 0; e < 10; e++ {
		go func() {
			f, e := i.GetResource()
			if e != nil {
				fmt.Println("error:" + e.Error())
			}

			f.(db).Select()

			i.Release(f)

			wg.Done()
		}()
	}

	wg.Wait()

}

type db struct {
	num int
}

func (d db) Close() error {
	fmt.Println("close db:" + strconv.Itoa(d.num))
	return nil
}

func (d db) Select() {
	fmt.Println("select something from db num=" + strconv.Itoa(d.num))
}
