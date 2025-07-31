package main

import (
	"acetek-mes/model"
	"math/rand"
	"time"

	"github.com/yxcloud1/go-comm/db"
)
var(
	spins[] model.SpinTraverse
	lines[] model.SpinLine
)

func init(){
	db.DB().Conn().Find(&spins)
	db.DB().Conn().Find(&lines)
	rand.Seed(time.Now().UnixNano())

}
func  spin(){
	for _, spin := range spins {
		rnd := rand.Intn(100000)
		spin.State = rnd
		rnd = rand.Intn(100000)
		if spin.IsBroken{
			spin.IsBroken = rnd < 200
		}else{
			spin.IsBroken = rnd < 1
		}
			go func(sp model.SpinTraverse) {
				if sp.SetTime == 0 {
					sp.CompletedTime = rand.Intn(280)
					sp.SetTime = 280
				}else{
					sp.CompletedTime = sp.CompletedTime + 1
				}
				if sp.State > 0 {
					sp.BeginTime = time.Now().Add(-time.Duration(sp.CompletedTime) * time.Minute)
					sp.EstEndTime = sp.BeginTime.Add(time.Duration(sp.CompletedTime + sp.CompensationTime) * time.Minute)
				}
				db.DB().Conn().Save(sp)
			}(spin)
		}

}