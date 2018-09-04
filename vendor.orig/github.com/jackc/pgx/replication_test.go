package pgx_test

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx"
)

// This function uses a postgresql 9.6 specific column
func getConfirmedFlushLsnFor(t *testing.T, conn *pgx.Conn, slot string) string {
	// Fetch the restart LSN of the slot, to establish a starting point
	rows, err := conn.Query(fmt.Sprintf("select confirmed_flush_lsn from pg_replication_slots where slot_name='%s'", slot))
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	defer rows.Close()

	var restartLsn string
	for rows.Next() {
		rows.Scan(&restartLsn)
	}
	return restartLsn
}

// This battleship test (at least somewhat by necessity) does
// several things all at once in a single run. It:
// - Establishes a replication connection & slot
// - Does a series of operations to create some known WAL entries
// - Replicates the entries down, and checks that the rows it
//   created come down in order
// - Sends a standby status message to update the server with the
//   wal position of the slot
// - Checks the wal position of the slot on the server to make sure
//   the update succeeded
func TestSimpleReplicationConnection(t *testing.T) {
	var err error

	if replicationConnConfig == nil {
		t.Skip("Skipping due to undefined replicationConnConfig")
	}

	conn := mustConnect(t, *replicationConnConfig)
	defer func() {
		// Ensure replication slot is destroyed, but don't check for errors as it
		// should have already been destroyed.
		conn.Exec("select pg_drop_replication_slot('pgx_test')")
		closeConn(t, conn)
	}()

	replicationConn := mustReplicationConnect(t, *replicationConnConfig)
	defer closeReplicationConn(t, replicationConn)

	var cp string
	var snapshot_name string
	cp, snapshot_name, err = replicationConn.CreateReplicationSlotEx("pgx_test", "test_decoding")
	if err != nil {
		t.Fatalf("replication slot create failed: %v", err)
	}
	if cp == "" {
		t.Logf("consistent_point is empty")
	}
	if snapshot_name == "" {
		t.Logf("snapshot_name is empty")
	}

	// Do a simple change so we can get some wal data
	_, err = conn.Exec("create table if not exists replication_test (a integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = replicationConn.StartReplication("pgx_test", 0, -1)
	if err != nil {
		t.Fatalf("Failed to start replication: %v", err)
	}

	var insertedTimes []int64
	currentTime := time.Now().Unix()

	for i := 0; i < 5; i++ {
		var ct pgx.CommandTag
		insertedTimes = append(insertedTimes, currentTime)
		ct, err = conn.Exec("insert into replication_test(a) values($1)", currentTime)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		t.Logf("Inserted %d rows", ct.RowsAffected())
		currentTime++
	}

	var foundTimes []int64
	var foundCount int
	var maxWal uint64

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	for {
		var message *pgx.ReplicationMessage

		message, err = replicationConn.WaitForReplicationMessage(ctx)
		if err != nil {
			t.Fatalf("Replication failed: %v %s", err, reflect.TypeOf(err))
		}

		if message.WalMessage != nil {
			// The waldata payload with the test_decoding plugin looks like:
			// public.replication_test: INSERT: a[integer]:2
			// What we wanna do here is check that once we find one of our inserted times,
			// that they occur in the wal stream in the order we executed them.
			walString := string(message.WalMessage.WalData)
			if strings.Contains(walString, "public.replication_test: INSERT") {
				stringParts := strings.Split(walString, ":")
				offset, err := strconv.ParseInt(stringParts[len(stringParts)-1], 10, 64)
				if err != nil {
					t.Fatalf("Failed to parse walString %s", walString)
				}
				if foundCount > 0 || offset == insertedTimes[0] {
					foundTimes = append(foundTimes, offset)
					foundCount++
				}
				if foundCount == len(insertedTimes) {
					break
				}
			}
			if message.WalMessage.WalStart > maxWal {
				maxWal = message.WalMessage.WalStart
			}

		}
		if message.ServerHeartbeat != nil {
			t.Logf("Got heartbeat: %s", message.ServerHeartbeat)
		}
	}

	for i := range insertedTimes {
		if foundTimes[i] != insertedTimes[i] {
			t.Fatalf("Found %d expected %d", foundTimes[i], insertedTimes[i])
		}
	}

	t.Logf("Found %d times, as expected", len(foundTimes))

	// Before closing our connection, let's send a standby status to update our wal
	// position, which should then be reflected if we fetch out our current wal position
	// for the slot
	status, err := pgx.NewStandbyStatus(maxWal)
	if err != nil {
		t.Errorf("Failed to create standby status %v", err)
	}
	replicationConn.SendStandbyStatus(status)

	restartLsn := getConfirmedFlushLsnFor(t, conn, "pgx_test")
	integerRestartLsn, _ := pgx.ParseLSN(restartLsn)
	if integerRestartLsn != maxWal {
		t.Fatalf("Wal offset update failed, expected %s found %s", pgx.FormatLSN(maxWal), restartLsn)
	}

	closeReplicationConn(t, replicationConn)

	replicationConn2 := mustReplicationConnect(t, *replicationConnConfig)
	defer closeReplicationConn(t, replicationConn2)

	err = replicationConn2.DropReplicationSlot("pgx_test")
	if err != nil {
		t.Fatalf("Failed to drop replication slot: %v", err)
	}

	droppedLsn := getConfirmedFlushLsnFor(t, conn, "pgx_test")
	if droppedLsn != "" {
		t.Errorf("Got odd flush lsn %s for supposedly dropped slot", droppedLsn)
	}
}

func TestReplicationConn_DropReplicationSlot(t *testing.T) {
	if replicationConnConfig == nil {
		t.Skip("Skipping due to undefined replicationConnConfig")
	}

	replicationConn := mustReplicationConnect(t, *replicationConnConfig)
	defer closeReplicationConn(t, replicationConn)

	var cp string
	var snapshot_name string
	cp, snapshot_name, err := replicationConn.CreateReplicationSlotEx("pgx_slot_test", "test_decoding")
	if err != nil {
		t.Logf("replication slot create failed: %v", err)
	}
	if cp == "" {
		t.Logf("consistent_point is empty")
	}
	if snapshot_name == "" {
		t.Logf("snapshot_name is empty")
	}

	err = replicationConn.DropReplicationSlot("pgx_slot_test")
	if err != nil {
		t.Fatalf("Failed to drop replication slot: %v", err)
	}

	// We re-create to ensure the drop worked.
	cp, snapshot_name, err = replicationConn.CreateReplicationSlotEx("pgx_slot_test", "test_decoding")
	if err != nil {
		t.Logf("replication slot create failed: %v", err)
	}
	if cp == "" {
		t.Logf("consistent_point is empty")
	}
	if snapshot_name == "" {
		t.Logf("snapshot_name is empty")
	}

	// And finally we drop to ensure we don't leave dirty state
	err = replicationConn.DropReplicationSlot("pgx_slot_test")
	if err != nil {
		t.Fatalf("Failed to drop replication slot: %v", err)
	}
}

func TestIdentifySystem(t *testing.T) {
	if replicationConnConfig == nil {
		t.Skip("Skipping due to undefined replicationConnConfig")
	}

	replicationConn2 := mustReplicationConnect(t, *replicationConnConfig)
	defer closeReplicationConn(t, replicationConn2)

	r, err := replicationConn2.IdentifySystem()
	if err != nil {
		t.Error(err)
	}
	defer r.Close()
	for _, fd := range r.FieldDescriptions() {
		t.Logf("Field: %s of type %v", fd.Name, fd.DataType)
	}

	var rowCount int
	for r.Next() {
		rowCount++
		values, err := r.Values()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Row values: %v", values)
	}
	if r.Err() != nil {
		t.Error(r.Err())
	}

	if rowCount == 0 {
		t.Errorf("Failed to find any rows: %d", rowCount)
	}
}

func getCurrentTimeline(t *testing.T, rc *pgx.ReplicationConn) int {
	r, err := rc.IdentifySystem()
	if err != nil {
		t.Error(err)
	}
	defer r.Close()
	for r.Next() {
		values, e := r.Values()
		if e != nil {
			t.Error(e)
		}
		return int(values[1].(int32))
	}
	t.Fatal("Failed to read timeline")
	return -1
}

func TestGetTimelineHistory(t *testing.T) {
	if replicationConnConfig == nil {
		t.Skip("Skipping due to undefined replicationConnConfig")
	}

	replicationConn := mustReplicationConnect(t, *replicationConnConfig)
	defer closeReplicationConn(t, replicationConn)

	timeline := getCurrentTimeline(t, replicationConn)

	r, err := replicationConn.TimelineHistory(timeline)
	if err != nil {
		t.Errorf("%#v", err)
	}
	defer r.Close()

	for _, fd := range r.FieldDescriptions() {
		t.Logf("Field: %s of type %v", fd.Name, fd.DataType)
	}

	var rowCount int
	for r.Next() {
		rowCount++
		values, err := r.Values()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Row values: %v", values)
	}
	if r.Err() != nil {
		if strings.Contains(r.Err().Error(), "No such file or directory") {
			// This is normal, this means the timeline we're on has no
			// history, which is the common case in a test db that
			// has only one timeline
			return
		}
		t.Error(r.Err())
	}

	// If we have a timeline history (see above) there should have been
	// rows emitted
	if rowCount == 0 {
		t.Errorf("Failed to find any rows: %d", rowCount)
	}
}

func TestStandbyStatusParsing(t *testing.T) {
	// Let's push the boundary conditions of the standby status and ensure it errors correctly
	status, err := pgx.NewStandbyStatus(0, 1, 2, 3, 4)
	if err == nil {
		t.Errorf("Expected error from new standby status, got %v", status)
	}

	// And if you provide 3 args, ensure the right fields are set
	status, err = pgx.NewStandbyStatus(1, 2, 3)
	if err != nil {
		t.Errorf("Failed to create test status: %v", err)
	}
	if status.WalFlushPosition != 1 {
		t.Errorf("Unexpected flush position %d", status.WalFlushPosition)
	}
	if status.WalApplyPosition != 2 {
		t.Errorf("Unexpected apply position %d", status.WalApplyPosition)
	}
	if status.WalWritePosition != 3 {
		t.Errorf("Unexpected write position %d", status.WalWritePosition)
	}
}
