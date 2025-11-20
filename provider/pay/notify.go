package pay

import (
	"github.com/asaskevich/EventBus"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func RegisterPayNotify(engine *gin.Engine, manager *PaymentManager) {
	engine.POST("/pay/notify/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		d, err := manager.GetDriver(provider)
		if err != nil {
			c.JSON(200, gin.H{"success": false, "msg": err.Error()})
			return
		}
		b, _ := ioutil.ReadAll(c.Request.Body)
		headers := map[string]string{}
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		ev, e := d.ParseNotify(c.Request.Context(), headers, b)
		if e != nil {
			c.JSON(200, gin.H{"success": false, "msg": e.Error()})
			return
		}
		if isDuplicate(ev.IdempotencyKey) {
			c.JSON(200, gin.H{"success": true, "data": ev})
			return
		}
		markKey(ev.IdempotencyKey)
		c.JSON(200, gin.H{"success": true, "data": ev})
	})
}

func RegisterPayNotifyWithBus(engine *gin.Engine, manager *PaymentManager, bus EventBus.Bus) {
	engine.POST("/pay/notify/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		d, err := manager.GetDriver(provider)
		if err != nil {
			c.JSON(200, gin.H{"success": false, "msg": err.Error()})
			return
		}
		b, _ := ioutil.ReadAll(c.Request.Body)
		headers := map[string]string{}
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		ev, e := d.ParseNotify(c.Request.Context(), headers, b)
		if e != nil {
			c.JSON(200, gin.H{"success": false, "msg": e.Error()})
			return
		}
		if isDuplicate(ev.IdempotencyKey) {
			c.JSON(200, gin.H{"success": true, "data": ev})
			return
		}
		markKey(ev.IdempotencyKey)
		bus.Publish("payment.notify", ev)
		c.JSON(200, gin.H{"success": true, "data": ev})
	})
}
