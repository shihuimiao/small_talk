package spool

import (
	"sync"
	"io"
	"fmt"
	"errors"
)

/**
    两点要注意的地方
	这个pool设置锁是为了close这个参数的线程不安全   所以close读写的时候要枷锁
	释放资源的地方都要调用io.Close接口的Close方法   为了释放资源
	代码里刚开始有一个地方忘了  看注视就知道是哪个地方
	close channle 和 for 循环的顺序要注意
 */

//定义一个pool
type pool struct {
	sync.Mutex
	res     chan io.Closer
	factory func() (io.Closer, error)
	close   bool
}

//创建一个pool
func New(f func() (io.Closer, error), size int) (*pool) {
	return &pool{
		res:     make(chan io.Closer, size),
		factory: f,
	}
}

//得到池子里的一个资源
func (p *pool) GetResource() (io.Closer, error) {
	select {
	case r, ok := <-p.res:
		if !ok {
			return nil, errors.New("pool is close")
		}
		fmt.Println("共享资源")
		return r, nil
	default:
		fmt.Println("新生成资源")
		return p.factory()
	}
}

//释放资源
func (p *pool) Release(f io.Closer) {
	//////忘了加锁    因为close是线程不安全的
	p.Lock()
	defer p.Unlock()

	if p.close {
		f.Close()
		return
	}

	select {
	case p.res <- f:
		fmt.Println("回收资源")
	default:
		fmt.Println("释放资源")
		///////这里忘了释放资源的操作了
		f.Close()
	}
}

//关闭资源池
func (p *pool) Close() {
	p.Lock()
	defer p.Unlock()

	if p.close {
		return
	}

	p.close = true

	///////这里for 循环和 close channel  顺序要调换

	//for v := range p.res {
	//	v.Close()
	//}
	//
	//close(p.res)

	//////这样是有好处的  先关闭   在循环   可以避免循环好后又有资源进入
	close(p.res)
	for v := range p.res {
		v.Close()
	}

}
