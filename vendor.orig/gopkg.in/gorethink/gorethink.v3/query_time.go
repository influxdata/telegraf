package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Now returns a time object representing the current time in UTC
func Now(args ...interface{}) Term {
	return constructRootTerm("Now", p.Term_NOW, args, map[string]interface{}{})
}

// Time creates a time object for a specific time
func Time(args ...interface{}) Term {
	return constructRootTerm("Time", p.Term_TIME, args, map[string]interface{}{})
}

// EpochTime returns a time object based on seconds since epoch
func EpochTime(args ...interface{}) Term {
	return constructRootTerm("EpochTime", p.Term_EPOCH_TIME, args, map[string]interface{}{})
}

// ISO8601Opts contains the optional arguments for the ISO8601 term
type ISO8601Opts struct {
	DefaultTimezone interface{} `gorethink:"default_timezone,omitempty"`
}

func (o ISO8601Opts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// ISO8601 returns a time object based on an ISO8601 formatted date-time string
func ISO8601(date interface{}, optArgs ...ISO8601Opts) Term {

	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("ISO8601", p.Term_ISO8601, []interface{}{date}, opts)
}

// InTimezone returns a new time object with a different time zone. While the
// time stays the same, the results returned by methods such as hours() will
// change since they take the timezone into account. The timezone argument
// has to be of the ISO 8601 format.
func (t Term) InTimezone(args ...interface{}) Term {
	return constructMethodTerm(t, "InTimezone", p.Term_IN_TIMEZONE, args, map[string]interface{}{})
}

// Timezone returns the timezone of the time object
func (t Term) Timezone(args ...interface{}) Term {
	return constructMethodTerm(t, "Timezone", p.Term_TIMEZONE, args, map[string]interface{}{})
}

// DuringOpts contains the optional arguments for the During term
type DuringOpts struct {
	LeftBound  interface{} `gorethink:"left_bound,omitempty"`
	RightBound interface{} `gorethink:"right_bound,omitempty"`
}

func (o DuringOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// During returns true if a time is between two other times
// (by default, inclusive for the start, exclusive for the end).
func (t Term) During(startTime, endTime interface{}, optArgs ...DuringOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "During", p.Term_DURING, []interface{}{startTime, endTime}, opts)
}

// Date returns a new time object only based on the day, month and year
// (ie. the same day at 00:00).
func (t Term) Date(args ...interface{}) Term {
	return constructMethodTerm(t, "Date", p.Term_DATE, args, map[string]interface{}{})
}

// TimeOfDay returns the number of seconds elapsed since the beginning of the
// day stored in the time object.
func (t Term) TimeOfDay(args ...interface{}) Term {
	return constructMethodTerm(t, "TimeOfDay", p.Term_TIME_OF_DAY, args, map[string]interface{}{})
}

// Year returns the year of a time object.
func (t Term) Year(args ...interface{}) Term {
	return constructMethodTerm(t, "Year", p.Term_YEAR, args, map[string]interface{}{})
}

// Month returns the month of a time object as a number between 1 and 12.
// For your convenience, the terms r.January(), r.February() etc. are
// defined and map to the appropriate integer.
func (t Term) Month(args ...interface{}) Term {
	return constructMethodTerm(t, "Month", p.Term_MONTH, args, map[string]interface{}{})
}

// Day return the day of a time object as a number between 1 and 31.
func (t Term) Day(args ...interface{}) Term {
	return constructMethodTerm(t, "Day", p.Term_DAY, args, map[string]interface{}{})
}

// DayOfWeek returns the day of week of a time object as a number between
// 1 and 7 (following ISO 8601 standard). For your convenience,
// the terms r.Monday(), r.Tuesday() etc. are defined and map to
// the appropriate integer.
func (t Term) DayOfWeek(args ...interface{}) Term {
	return constructMethodTerm(t, "DayOfWeek", p.Term_DAY_OF_WEEK, args, map[string]interface{}{})
}

// DayOfYear returns the day of the year of a time object as a number between
// 1 and 366 (following ISO 8601 standard).
func (t Term) DayOfYear(args ...interface{}) Term {
	return constructMethodTerm(t, "DayOfYear", p.Term_DAY_OF_YEAR, args, map[string]interface{}{})
}

// Hours returns the hour in a time object as a number between 0 and 23.
func (t Term) Hours(args ...interface{}) Term {
	return constructMethodTerm(t, "Hours", p.Term_HOURS, args, map[string]interface{}{})
}

// Minutes returns the minute in a time object as a number between 0 and 59.
func (t Term) Minutes(args ...interface{}) Term {
	return constructMethodTerm(t, "Minutes", p.Term_MINUTES, args, map[string]interface{}{})
}

// Seconds returns the seconds in a time object as a number between 0 and
// 59.999 (double precision).
func (t Term) Seconds(args ...interface{}) Term {
	return constructMethodTerm(t, "Seconds", p.Term_SECONDS, args, map[string]interface{}{})
}

// ToISO8601 converts a time object to its iso 8601 format.
func (t Term) ToISO8601(args ...interface{}) Term {
	return constructMethodTerm(t, "ToISO8601", p.Term_TO_ISO8601, args, map[string]interface{}{})
}

// ToEpochTime converts a time object to its epoch time.
func (t Term) ToEpochTime(args ...interface{}) Term {
	return constructMethodTerm(t, "ToEpochTime", p.Term_TO_EPOCH_TIME, args, map[string]interface{}{})
}

var (
	// Days

	// Monday is a constant representing the day of the week Monday
	Monday = constructRootTerm("Monday", p.Term_MONDAY, []interface{}{}, map[string]interface{}{})
	// Tuesday is a constant representing the day of the week Tuesday
	Tuesday = constructRootTerm("Tuesday", p.Term_TUESDAY, []interface{}{}, map[string]interface{}{})
	// Wednesday is a constant representing the day of the week Wednesday
	Wednesday = constructRootTerm("Wednesday", p.Term_WEDNESDAY, []interface{}{}, map[string]interface{}{})
	// Thursday is a constant representing the day of the week Thursday
	Thursday = constructRootTerm("Thursday", p.Term_THURSDAY, []interface{}{}, map[string]interface{}{})
	// Friday is a constant representing the day of the week Friday
	Friday = constructRootTerm("Friday", p.Term_FRIDAY, []interface{}{}, map[string]interface{}{})
	// Saturday is a constant representing the day of the week Saturday
	Saturday = constructRootTerm("Saturday", p.Term_SATURDAY, []interface{}{}, map[string]interface{}{})
	// Sunday is a constant representing the day of the week Sunday
	Sunday = constructRootTerm("Sunday", p.Term_SUNDAY, []interface{}{}, map[string]interface{}{})

	// Months

	// January is a constant representing the month January
	January = constructRootTerm("January", p.Term_JANUARY, []interface{}{}, map[string]interface{}{})
	// February is a constant representing the month February
	February = constructRootTerm("February", p.Term_FEBRUARY, []interface{}{}, map[string]interface{}{})
	// March is a constant representing the month March
	March = constructRootTerm("March", p.Term_MARCH, []interface{}{}, map[string]interface{}{})
	// April is a constant representing the month April
	April = constructRootTerm("April", p.Term_APRIL, []interface{}{}, map[string]interface{}{})
	// May is a constant representing the month May
	May = constructRootTerm("May", p.Term_MAY, []interface{}{}, map[string]interface{}{})
	// June is a constant representing the month June
	June = constructRootTerm("June", p.Term_JUNE, []interface{}{}, map[string]interface{}{})
	// July is a constant representing the month July
	July = constructRootTerm("July", p.Term_JULY, []interface{}{}, map[string]interface{}{})
	// August is a constant representing the month August
	August = constructRootTerm("August", p.Term_AUGUST, []interface{}{}, map[string]interface{}{})
	// September is a constant representing the month September
	September = constructRootTerm("September", p.Term_SEPTEMBER, []interface{}{}, map[string]interface{}{})
	// October is a constant representing the month October
	October = constructRootTerm("October", p.Term_OCTOBER, []interface{}{}, map[string]interface{}{})
	// November is a constant representing the month November
	November = constructRootTerm("November", p.Term_NOVEMBER, []interface{}{}, map[string]interface{}{})
	// December is a constant representing the month December
	December = constructRootTerm("December", p.Term_DECEMBER, []interface{}{}, map[string]interface{}{})
)
