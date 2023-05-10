package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const (
	N = 100000 // 数组长度
	M = 5      // worker数量
	R = 10000  // 重复次数
	T = 10     // 超时秒数
)

var (
	S    [N]int          // 整数数组
	lock [N]sync.RWMutex // 读写锁数组
	wg   sync.WaitGroup  // 等待组
)

func main() {
	// 初始化数组
	for i := 0; i < N; i++ {
		S[i] = rand.Intn(100)
	}

	// 启动M个worker并发访问和更新数组
	wg.Add(M)
	for i := 0; i < M; i++ {
		go worker()
	}

	// 等待所有worker完成
	wg.Wait()

	// 打印数组前10个元素
	fmt.Println(S[:10])
}

func worker() {
	defer wg.Done()
	for r := 0; r < R; r++ {
		// 随机生成i,j
		i := rand.Intn(N)
		j := rand.Intn(N)

		// 创建一个缓冲通道用于发送和接收数据
		ch := make(chan int, 4)

		// 创建一个超时通道用于取消阻塞操作
		timeout := time.After(time.Duration(T) * time.Second)

		// 尝试获取所有需要的锁
		go func() {
			// 对需要获取的锁进行排序，按照升序获取和释放锁
			indexes := []int{i, (i + 1) % N, (i + 2) % N, j}
			sort.Ints(indexes)
			for _, k := range indexes {
				if k == j {
					lock[k].Lock()
				} else {
					lock[k].RLock()
				}
				defer func(k int) {
					if k == j {
						lock[k].Unlock()
					} else {
						lock[k].RUnlock()
					}
				}(k)
			}

			ch <- S[i]
			ch <- S[(i+1)%N]
			ch <- S[(i+2)%N]

			// 如果j在[i,i+2]区间，先保存S(j)的值
			if i <= j && j <= (i+2)%N {
				ch <- S[j]
			}

			close(ch)
		}()

		var a, b, c, d int

		select { // 如果成功获取所有的锁，就进行操作
		case a = <-ch:
			b = <-ch
			c = <-ch

			// 更新S(i)为S(i) + S(i+1) + S(i+2)
			S[i] = a + b + c

			// 如果j在[i,i+2]区间，再更新S(j)为原来的值
			if i <= j && j <= (i+2)%N {
				d = <-ch
				S[j] = d
			}

		// 如果超时，就放弃操作
		case <-timeout:
			fmt.Println("Timeout!")
		}
	}
}
