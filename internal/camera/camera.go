package camera

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type IpCamera struct {
	Address    string
	Username   string
	Password   string
	lightlevel int
}

func (c *IpCamera) CurrentLightLevel() int {
	return c.lightlevel
}

func (c *IpCamera) IncreaseLight() {
	if c.lightlevel < 10 {
		err := c.sendLightAdjustRequest("4")
		if err != nil {
			return
		}
		c.lightlevel++
	}
}

func (c *IpCamera) DecreaseLight() {
	if c.lightlevel > 0 {
		err := c.sendLightAdjustRequest("5")
		if err != nil {
			return
		}
		c.lightlevel--
	}
}
func (c *IpCamera) ZeroLight() {
	for i := 0; i < 10; i++ {
        err := c.sendLightAdjustRequest("5")
        if(err != nil){
            log.Printf("Could not zero light: %v", err)
        }
	}

	c.lightlevel = 0
}

func (c *IpCamera) SetLightLevel(l int) error {
	if l > 10 {
		return fmt.Errorf("Light level max 10")
	}

	if l < 0 {
		return fmt.Errorf("Light level min 0")
	}

	if c.lightlevel > l {
		log.Printf("Lowering light level to %d\n", l)
		for ; c.lightlevel > l; c.lightlevel-- {
			log.Printf("%d, target: %d\n", c.lightlevel, l)
			c.DecreaseLight()
		}
		return nil
	}

	if c.lightlevel < l {
		log.Printf("Raising light level to %d\n", l)
		for ; c.lightlevel < l; c.lightlevel++ {
			log.Printf("%d, target: %d\n", c.lightlevel, l)
			c.IncreaseLight()
		}
		return nil
	}

	return nil
}

func (c *IpCamera) sendLightAdjustRequest(actionCode string) error {

	// Generating the MD5 hash for the username and password
	auth := fmt.Sprintf("%x", md5.Sum([]byte(c.Username+":"+c.Password)))

	// The data to be sent in the POST request
	params := fmt.Sprintf("?action=update&group=PTZCTRL&channel=0&PTZCTRL.action=%s&PTZCTRL.speed=50&nRanId=%s", actionCode, randomString(8))

	url := fmt.Sprintf("http://%s/cgi-bin/control.cgi", c.Address)

	// Creating the request
	req, err := http.NewRequest("POST", url+params, nil)
	if err != nil {
		panic(err)
	}

	// Setting the content type and the authorization headers
	req.Header.Set("Content-Type", "text/html; charset=UTF-8")
	req.Header.Set("Authorization", "Md5 "+auth)

	// Making the request
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	client.Do(req)

	return nil

}

// randomString generates a random string of n characters
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
