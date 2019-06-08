package global

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)
//This package is not thread safe
type Status struct {
	lastStatus map[string]string
	CurrStatus map[string]string
	currTime   time.Time
	lastTime   time.Time
	servertag  string
	ok         bool
}

func New(servertag string) *Status {
	g := &Status{}
	g.CurrStatus = nil
	g.lastStatus = nil
	g.lastTime = time.Now()
	g.currTime = time.Now()
	g.ok = true
	g.servertag = servertag

	return g
}

func (g *Status) Fill( db * sql.DB ) error {
	defer func() {
		if ! g.ok {
			//clean history data if current fetch failed
			// otherwise the delta is not expected since they will cross multi intervals
			g.lastStatus = nil
			g.CurrStatus = nil
		}
	}()

	g.ok = false
	currTime := time.Now()

	rows, err := db.Query("SHOW global status")
	if err != nil {
		return err
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key string
		var val string

		if err = rows.Scan(&key, &val); err != nil {
			return err
		}

		values[key] = val
	}

	if g.CurrStatus != nil {
		g.lastStatus = g.CurrStatus
		g.lastTime = g.currTime
	}

	g.CurrStatus = values
	g.currTime =  currTime

	g.ok = true
	return nil
}

func (g *Status) GetProperty(property string) ( string, error) {
	if  g.CurrStatus == nil  {
		return "", fmt.Errorf("errror getting [%s] property: CurrStatus is nil", g.servertag)
	}

	val, ok := g.CurrStatus[property]
	if ! ok {
		return "", fmt.Errorf("errror getting [%s] property %s doesnot exist", g.servertag, property)
	}

	return string(val), nil
}

func (g *Status) GetPropertyDelta( property string ) (int64, error) {
	if g.lastStatus == nil {
		return  0, fmt.Errorf("error getting [%s] propery delta value, property: %s, no history data yet", g.servertag, property)
	}

	if  g.CurrStatus == nil  {
		return 0, fmt.Errorf("errror getting [%s] property: CurrStatus is nil", g.servertag)
	}

	lastVal, ok := g.CurrStatus[property]
	if ! ok {
		return 0, fmt.Errorf("errror getting [%s] property delta, history property  %s doesnot exist", g.servertag, property)
	}

	currVal, ok := g.CurrStatus[property]
	if ! ok {
		return 0, fmt.Errorf("errror getting [%s] property delta, property  %s doesnot exist", g.servertag, property)
	}

	lastNum, err := strconv.ParseInt(lastVal, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("errror getting [%s] property delta, property  %s is not a number: %s", g.servertag, property, err)
	}

	currNum, err := strconv.ParseInt(currVal, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("errror getting [%s] property delta, property  %s is not a number: %s", g.servertag, property, err)
	}

	return currNum - lastNum, nil
}
