// +build windows
package ping

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/influxdata/telegraf/testutil"
)

// Windows ping format ( should support multilanguage ?)
var winPLPingOutput = `
Badanie 8.8.8.8 z 32 bajtami danych:
Odpowiedz z 8.8.8.8: bajtow=32 czas=49ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=46ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=48ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=57ms TTL=43

Statystyka badania ping dla 8.8.8.8:
    Pakiety: Wyslane = 4, Odebrane = 4, Utracone = 0
             (0% straty),
Szacunkowy czas bladzenia pakietww w millisekundach:
    Minimum = 46 ms, Maksimum = 57 ms, Czas sredni = 50 ms
`
// Windows ping format ( should support multilanguage ?)
var winENPingOutput = `
Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=52ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=51ms TTL=43

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 50ms, Maximum = 52ms, Average = 50ms
`

func TestHost( t* testing.T ) {
	trans, rec, avg, min, max, err := processPingOutput(winPLPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, rec, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 46, min, "Min 46")
	assert.Equal(t, 57, max, "max 57")
	
	trans, rec, avg, min, max, err = processPingOutput(winENPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, rec, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 50, min, "Min 50")
	assert.Equal(t, 52, max, "Max 52")
}

func mockHostPinger(timeout float64, args ...string) (string, error) {
	return winENPingOutput, nil
}

// Test that Gather function works on a normal ping
func TestPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com", "www.reddit.com"},
		pingHost: mockHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    4,
		"percent_packet_loss": 0.0,
		"average_response_ms": 50,
		"minimum_response_ms": 50,
		"maximum_response_ms": 52,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	tags = map[string]string{"url": "www.reddit.com"}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var errorPingOutput = `
Badanie nask.pl [195.187.242.157] z 32 bajtami danych:
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.

Statystyka badania ping dla 195.187.242.157:
    Pakiety: Wysłane = 4, Odebrane = 0, Utracone = 4
             (100% straty),
`

func mockErrorHostPinger(timeout float64, args ...string) (string, error) {
	return errorPingOutput, errors.New("No packets received")
}

// Test that Gather works on a ping with no transmitted packets, even though the
// command returns an error
func TestBadPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockErrorHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    0,
		"percent_packet_loss": 100.0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var lossyPingOutput = `
Badanie thecodinglove.com [66.6.44.4] z 9800 bajtami danych:
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=118ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=119ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=116ms TTL=48

Statystyka badania ping dla 66.6.44.4:
    Pakiety: Wysłane = 9, Odebrane = 7, Utracone = 2
             (22% straty),
Szacunkowy czas błądzenia pakietów w millisekundach:
    Minimum = 114 ms, Maksimum = 119 ms, Czas średni = 115 ms
`

func mockLossyHostPinger(timeout float64, args ...string) (string, error) {
	return lossyPingOutput, nil
}

// Test that Gather works on a ping with lossy packets
func TestLossyPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockLossyHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 9,
		"packets_received":    7,
		"percent_packet_loss": 22.22222222222222,
		"average_response_ms": 115,
		"minimum_response_ms": 114,
		"maximum_response_ms": 119,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

// Fatal ping output (invalid argument)
var fatalPingOutput = `
Zla opcja -d.


Sposob uzycia: ping [-t] [-a] [-n liczba] [-l rozmiar] [-f] [-i TTL] [-v TOS]
               [-r liczba] [-s liczba] [[-j lista_hostow] | [-k lista_hostow]]
               [-w limit_czasu] [-R] [-S adres_zrodlowy] [-4] [-6]
               nazwa_obiektu_docelowego
Opcje:
    -t               Odpytuje okreslonego hosta do czasu zatrzymania.
                     Aby przejrzec statystyki i kontynuowac,
                     nacisnij klawisze Ctrl+Break.
                     Aby zakonczyc, nacisnij klawisze Ctrl+C.
   -a                Tlumaczy adresy na nazwy hostow.
   -n liczba         Liczba wysylanych ządan echa.
   -l rozmiar        Rozmiar buforu wysylania.
   -f                Ustawia w pakiecie flagę "Nie fragmentuj" (tylko IPv4).
   -i TTL            Czas wygasnięcia.
   -v TOS            Typ uslugi (tylko IPv4). To ustawienie zostalo
                     zaniechane i nie ma wplywu na wartosc pola typu uslugi
                     w naglowku IP.
   -r liczba         Rejestruje trasę dla podanej liczby przeskokow (tylko IPv4).
   -s liczba         Sygnatura czasowa dla podanej liczby przeskokow (tylko IPv4).
   -j lista_hostow   Swobodna trasa zrodlowa wg listy lista_hostow
                     (tylko IPv4).
   -k lista_hostow   scisle okreslona trasa zrodlowa wg listy lista_hostow
                     (tylko IPv4).
   -w limit_czasu    Limit czasu oczekiwania na odpowiedz (w  milisekundach).
   -R                Powoduje uzycie naglowka routingu w celu dodatkowego
                     testowania trasy wstecznej (tylko IPv6).
   -S adres_zrodlowy Adres zrodlowy do uzycia.
   -4                Wymusza uzywanie IPv4.
   -6                Wymusza uzywanie IPv6.
`

func mockFatalHostPinger(timeout float64, args ...string) (string, error) {
	return fatalPingOutput, errors.New("So very bad")
}

// Test that a fatal ping command does not gather any statistics.
func TestFatalPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockFatalHostPinger,
	}

	p.Gather(&acc)
	assert.False(t, acc.HasMeasurement("packets_transmitted"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("packets_received"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("average_response_ms"),
		"Fatal ping should not have packet measurements")
}
