package public

import (
	"fmt"
	"github.com/e421083458/go_gateway/golang_common/lib"
	"github.com/garyburd/redigo/redis"
	"sync/atomic"
	"time"
)

type RedisFlowCountService struct {
	AppID       string
	Interval    time.Duration
	QPS         int64	// 两次Ticker之间的qps
	Unix        int64
	TickerCount int64	// 两次ticker之间的请求数量缓存，每次ticker后会清空。
	TotalCount  int64	// 上次ticker时，当天的总量
}

func NewRedisFlowCountService(appID string, interval time.Duration) *RedisFlowCountService {
	reqCounter := &RedisFlowCountService{
		AppID:    appID,
		Interval: interval,
		QPS:      0,
		Unix:     0,
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		ticker := time.NewTicker(interval)
		for {
			// 定期将数据刷到redis，并重置本地数据。
			<-ticker.C

			// 重置本地数据，需要原子操作。
			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount) //获取数据
			atomic.StoreInt64(&reqCounter.TickerCount, 0)            //重置数据

			// 叠加到当天，当前小时
			currentTime := time.Now()
			dayKey := reqCounter.GetDayKey(currentTime)
			hourKey := reqCounter.GetHourKey(currentTime)
			if err := RedisConfPipline(func(c redis.Conn) {
				c.Send("INCRBY", dayKey, tickerCount)
				c.Send("EXPIRE", dayKey, 86400*2) // 过期时间，2天
				c.Send("INCRBY", hourKey, tickerCount)
				c.Send("EXPIRE", hourKey, 86400*2) // 过期时间，2天
			}); err != nil {
				fmt.Println("RedisConfPipline err",err)
				continue
			}

			// 查询一下，当前的总量
			totalCount, err := reqCounter.GetDayData(currentTime)
			if err != nil {
				fmt.Println("reqCounter.GetDayData err",err)
				continue
			}
			nowUnix := time.Now().Unix()
			if reqCounter.Unix == 0 {
				reqCounter.Unix = time.Now().Unix()
				continue
			}
			tickerCount = totalCount - reqCounter.TotalCount // 本次ticker的当天总量 - 上次ticker的当天总量
			if nowUnix > reqCounter.Unix {
				reqCounter.TotalCount = totalCount // 保存本次ticker的当天总量，为了下次ticker使用
				reqCounter.QPS = tickerCount / (nowUnix - reqCounter.Unix) // 最近两次ticker的差值 / 时间差 = 两次ticker之间的qps
				reqCounter.Unix = time.Now().Unix()
			}
		}
	}()
	return reqCounter
}

func (o *RedisFlowCountService) GetDayKey(t time.Time) string {
	dayStr := t.In(lib.TimeLocation).Format("20060102")
	return fmt.Sprintf("%s_%s_%s", RedisFlowDayKey, dayStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourKey(t time.Time) string {
	hourStr := t.In(lib.TimeLocation).Format("2006010215")
	return fmt.Sprintf("%s_%s_%s", RedisFlowHourKey, hourStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourData(t time.Time) (int64,error) {
	return redis.Int64(RedisConfDo("GET", o.GetHourKey(t)))
}

func (o *RedisFlowCountService) GetDayData(t time.Time) (int64,error) {
	return redis.Int64(RedisConfDo("GET", o.GetDayKey(t)))
}

//原子增加
func (o *RedisFlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&o.TickerCount, 1)
	}()
}
