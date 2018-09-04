package kazoo

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

func TestConsumergroups(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroups")

	cgs, err := kz.Consumergroups()
	if err != nil {
		t.Error(err)
	}
	originalCount := len(cgs)

	if cg := cgs.Find(cg.Name); cg != nil {
		t.Error("Consumergoup `test.kazoo.TestConsumergroups` should not be found")
	}

	if exists, _ := cg.Exists(); exists {
		t.Error("Consumergoup `test.kazoo.TestConsumergroups` should not be registered yet")
	}

	if err := cg.Create(); err != nil {
		t.Error(err)
	}

	if exists, _ := cg.Exists(); !exists {
		t.Error("Consumergoup `test.kazoo.TestConsumergroups` should be registered now")
	}

	cgs, err = kz.Consumergroups()
	if err != nil {
		t.Error(err)
	}

	if len(cgs) != originalCount+1 {
		t.Error("Should have one more consumergroup than at the start")
	}

	if err := cg.Delete(); err != nil {
		t.Error(err)
	}

	if exists, _ := cg.Exists(); exists {
		t.Error("Consumergoup `test.kazoo.TestConsumergroups` should not be registered anymore")
	}

	cgs, err = kz.Consumergroups()
	if err != nil {
		t.Error(err)
	}

	if len(cgs) != originalCount {
		t.Error("Should have the original number of consumergroups again")
	}
}

func TestConsumergroupInstances(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupInstances")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	if instances, err := cg.Instances(); err != nil {
		t.Error(err)
	} else if len(instances) != 0 {
		t.Fatal("Expected no active consumergroup instances")
	}

	instance1 := cg.NewInstance()
	// Make sure that the instance is unregistered.
	if reg, err := instance1.Registration(); err != ErrInstanceNotRegistered || reg != nil {
		t.Errorf("Expected no registration: reg=%v, err=(%v)", reg, err)
	}

	// Register a new instance
	if instance1.ID == "" {
		t.Error("It should generate a valid instance ID")
	}
	if err := instance1.Register([]string{"topic"}); err != nil {
		t.Error(err)
	}

	// Verify registration
	reg, err := instance1.Registration()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(reg.Subscription, map[string]int{"topic": 1}) {
		t.Errorf("Unexpected registration: %v", reg)
	}

	// Try to register an instance with the same ID.
	if err := cg.Instance(instance1.ID).Register([]string{"topic"}); err != ErrInstanceAlreadyRegistered {
		t.Error("The instance should already be registered")
	}

	instance2 := cg.Instance("test")
	if err := instance2.Register([]string{"topic"}); err != nil {
		t.Error(err)
	}

	time.Sleep(50 * time.Millisecond)

	if instances, err := cg.Instances(); err != nil {
		t.Error(err)
	} else {
		if len(instances) != 2 {
			t.Error("Expected 2 active consumergroup instances")
		}
		if i := instances.Find(instance1.ID); i == nil {
			t.Error("Expected instance1 to be registered.")
		}
		if i := instances.Find(instance2.ID); i == nil {
			t.Error("Expected instance2 to be registered.")
		}
	}

	// Deregister the two running instances
	if err := instance1.Deregister(); err != nil {
		t.Error(err)
	}
	if err := instance2.Deregister(); err != nil {
		t.Error(err)
	}

	// Try to deregister an instance that was not register
	instance3 := cg.NewInstance()
	if err := instance3.Deregister(); err != ErrInstanceNotRegistered {
		t.Error("Expected new instance to not be registered")
	}

	if instances, err := cg.Instances(); err != nil {
		t.Error(err)
	} else if len(instances) != 0 {
		t.Error("Expected no active consumergroup instances")
	}
}

func TestConsumergroupInstanceCrash(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupInstancesEphemeral")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	// Create a kazoo instance on which we will simulate a crash.
	config := NewConfig()
	config.Timeout = 50 * time.Millisecond
	crashingKazoo, err := NewKazoo(zookeeperPeers, config)
	if err != nil {
		t.Fatal(err)
	}
	crashingCG := crashingKazoo.Consumergroup(cg.Name)

	// Instantiate and register the instance.
	instance := crashingCG.NewInstance()
	if err := instance.Register([]string{"test.1"}); err != nil {
		t.Error(err)
	}

	time.Sleep(50 * time.Millisecond)
	if instances, err := cg.Instances(); err != nil {
		t.Error(err)
	} else if len(instances) != 1 {
		t.Error("Should have 1 running instance, found", len(instances))
	}

	// Simulate a crash, and wait for Zookeeper to pick it up
	_ = crashingKazoo.Close()
	time.Sleep(200 * time.Millisecond)

	if instances, err := cg.Instances(); err != nil {
		t.Error(err)
	} else if len(instances) != 0 {
		t.Error("Should have 0 running instances")
	}
}

func TestConsumergroupWatchInstances(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupWatchInstances")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	instances, c, err := cg.WatchInstances()
	if err != nil {
		t.Fatal(err)
	}

	if len(instances) != 0 {
		t.Error("Expected 0 running instances")
	}

	instance := cg.NewInstance()

	// Make sure a proper error is returned when an unregistered instance is
	// updated.
	err = instance.UpdateRegistration([]string{"foo"})
	if err != ErrInstanceNotRegistered {
		t.Fatal("Expected ErrInstanceNotRegistered")
	}

	if err := instance.Register([]string{"topic"}); err != nil {
		t.Fatal(err)
	}

	// The instance watch should have been triggered
	assertWatchTriggered(t, c, 5*time.Second)

	instances, c, err = cg.WatchInstances()
	if err != nil {
		t.Fatal(err)
	}

	if len(instances) != 1 {
		t.Error("Expected 1 running instance")
	}

	if err := instance.Deregister(); err != nil {
		t.Fatal(err)
	}

	// The instance watch should have been triggered again
	assertWatchTriggered(t, c, 5*time.Second)

	instances, err = cg.Instances()
	if err != nil {
		t.Fatal(err)
	}

	if len(instances) != 0 {
		t.Error("Expected 0 running instances")
	}
}

func TestConsumergroupInstanceWatchRegistration(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupWatchRegistration")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	instance := cg.NewInstance()
	if err := instance.Register([]string{"topic"}); err != nil {
		t.Fatal(err)
	}

	// Set a watch and make sure that it does not trigger on its own.
	reg, c, err := instance.WatchRegistration()
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(reg.Subscription, map[string]int{}) {
		t.Fatalf("Bad subscription, got=%v", reg.Subscription)
	}
	assertWatchNotTriggered(t, c, 200*time.Millisecond)

	// The watch is triggered by an update.
	if err = instance.UpdateRegistration([]string{"foo", "bar"}); err != nil {
		t.Fatal(err)
	}
	assertWatchTriggered(t, c, 5*time.Second)

	// Update registration.
	if err = instance.UpdateRegistration([]string{"foo", "bazz"}); err != nil {
		t.Fatal(err)
	}

	// Set a watch again.
	reg, c, err = instance.WatchRegistration()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reg.Subscription, map[string]int{"foo": 1, "bazz": 1}) {
		t.Fatalf("Bad subscription, got=%v", reg.Subscription)
	}
	assertWatchNotTriggered(t, c, 200*time.Millisecond)

	// A watch is triggered by any update operation even if the value does not
	// change.
	if err = instance.UpdateRegistration([]string{"foo", "bazz"}); err != nil {
		t.Fatal(err)
	}
	assertWatchTriggered(t, c, 5*time.Second)

	// Cleanup
	if err := instance.Deregister(); err != nil {
		t.Fatal(err)
	}
}

func TestConsumergroupInstanceClaimPartition(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupInstanceClaimPartition")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	// Create two instances for this consumergroup

	i1 := cg.NewInstance()
	if err := i1.Register([]string{"test.4"}); err != nil {
		t.Fatal(err)
	}
	i2 := cg.NewInstance()
	if err := i2.Register([]string{"test.4"}); err != nil {
		t.Fatal(err)
	}

	// Claim all partitions divided by instance 1 and 2

	if err := i1.ClaimPartition("test.4", 0); err != nil {
		t.Error(err)
	}
	if err := i1.ClaimPartition("test.4", 1); err != nil {
		t.Error(err)
	}
	if err := i2.ClaimPartition("test.4", 2); err != nil {
		t.Error(err)
	}
	if err := i2.ClaimPartition("test.4", 3); err != nil {
		t.Error(err)
	}

	// Try to claim more partitions
	if err := i1.ClaimPartition("test.4", 3); err != ErrPartitionClaimedByOther {
		t.Error("Expected ErrPartitionClaimedByOther to be returned, found", err)
	}

	if err := i2.ClaimPartition("test.4", 0); err != ErrPartitionClaimedByOther {
		t.Error("Expected ErrPartitionClaimedByOther to be returned, found", err)
	}

	// Instance 1: release some partitions

	if err := i1.ReleasePartition("test.4", 0); err != nil {
		t.Error(err)
	}
	if err := i1.ReleasePartition("test.4", 1); err != nil {
		t.Error(err)
	}

	// Instance 2: claim the released partitions

	if err := i2.ClaimPartition("test.4", 0); err != nil {
		t.Error(err)
	}
	if err := i2.ClaimPartition("test.4", 1); err != nil {
		t.Error(err)
	}

	// Instance 2: release all partitions

	if err := i2.ReleasePartition("test.4", 0); err != nil {
		t.Error(err)
	}
	if err := i2.ReleasePartition("test.4", 1); err != nil {
		t.Error(err)
	}
	if err := i2.ReleasePartition("test.4", 2); err != nil {
		t.Error(err)
	}
	if err := i2.ReleasePartition("test.4", 3); err != nil {
		t.Error(err)
	}

	if err := i1.Deregister(); err != nil {
		t.Error(err)
	}
	if err := i2.Deregister(); err != nil {
		t.Error(err)
	}
}

func TestConsumergroupInstanceClaimPartitionSame(t *testing.T) {
	// Given
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupInstanceClaimPartition2")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	instance := cg.NewInstance()
	if err := instance.Register([]string{"test.4"}); err != nil {
		t.Fatal(err)
	}

	if err := instance.ClaimPartition("test.4", 0); err != nil {
		t.Error(err)
	}

	// When: claim the same partition again
	err = instance.ClaimPartition("test.4", 0)

	// Then
	if err != nil {
		t.Error(err)
	}

	// Cleanup
	if err := instance.ReleasePartition("test.4", 0); err != nil {
		t.Error(err)
	}
	if err := instance.Deregister(); err != nil {
		t.Error(err)
	}
}

func TestConsumergroupInstanceWatchPartitionClaim(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupInstanceWatchPartitionClaim")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	instance1 := cg.NewInstance()
	if err := instance1.Register([]string{"test.4"}); err != nil {
		t.Fatal(err)
	}

	// Assert the partition isn't claimed
	instance, change, err := cg.WatchPartitionOwner("test.4", 0)
	if err != nil {
		t.Fatal(err)
	}
	if instance != nil {
		t.Fatal("An unclaimed partition should not return an instance")
	}
	if change != nil {
		t.Fatal("An unclaimed partition should not return a watch")
	}

	// Now claim the partition
	if err := instance1.ClaimPartition("test.4", 0); err != nil {
		t.Fatal(err)
	}

	// This time, we should get an insance back
	instance, change, err = cg.WatchPartitionOwner("test.4", 0)
	if err != nil {
		t.Fatal(err)
	}

	if instance.ID != instance1.ID {
		t.Error("Our instance should have claimed the partition")
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := instance1.ReleasePartition("test.4", 0); err != nil {
			t.Fatal(err)
		}
	}()

	// Wait for the zookeeper watch to trigger
	<-change

	// Ensure the partition is no longer claimed
	instance, err = cg.PartitionOwner("test.4", 0)
	if err != nil {
		t.Fatal(err)
	}
	if instance != nil {
		t.Error("The partition should have been release by now")
	}

	// Cleanup
	if err := instance1.Deregister(); err != nil {
		t.Error(err)
	}
}

func TestConsumergroupOffsets(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupOffsets")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	offset, err := cg.FetchOffset("test", 0)
	if err != nil {
		t.Error(err)
	}

	if offset >= 0 {
		t.Error("Expected to get a negative offset for a partition that hasn't seen an offset commit yet")
	}

	if err := cg.CommitOffset("test", 0, 1234); err != nil {
		t.Error(err)
	}

	offset, err = cg.FetchOffset("test", 0)
	if err != nil {
		t.Error(err)
	}
	if offset != 1234 {
		t.Error("Expected to get the offset that was committed.")
	}
}

func TestConsumergroupResetOffsetsRace(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupResetOffsetsRace")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	offsets, err := cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if len(offsets) > 0 {
		t.Errorf("A new consumergroup shouldn't have any offsets set, but found offsets for %d topics", len(offsets))
	}

	if err := cg.CommitOffset("test", 0, 1234); err != nil {
		t.Error(err)
	}

	if err := cg.CommitOffset("test", 1, 2345); err != nil {
		t.Error(err)
	}

	offsets, err = cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if offsets["test"][0] == 1234 && offsets["test"][1] == 2345 {
		t.Log("All offsets present in offset map")
	} else {
		t.Logf("Offset map not as expected: %v", offsets)
	}

	cg2 := kz.Consumergroup("test.kazoo.TestConsumergroupResetOffsetsRace")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := cg2.ResetOffsets(); err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := cg.ResetOffsets(); err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	offsets, err = cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if len(offsets) > 0 {
		t.Errorf("After a reset, consumergroup shouldn't have any offsets set, but found offsets for %d topics", len(offsets))
	}
}

func TestConsumergroupResetOffsets(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer assertSuccessfulClose(t, kz)

	cg := kz.Consumergroup("test.kazoo.TestConsumergroupResetOffsets")
	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cg.Delete(); err != nil {
			t.Error(err)
		}
	}()

	offsets, err := cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if len(offsets) > 0 {
		t.Errorf("A new consumergroup shouldn't have any offsets set, but found offsets for %d topics", len(offsets))
	}

	if err := cg.CommitOffset("test1", 0, 1234); err != nil {
		t.Error(err)
	}

	if err := cg.CommitOffset("test1", 1, 2345); err != nil {
		t.Error(err)
	}

	if err := cg.CommitOffset("test2", 0, 3456); err != nil {
		t.Error(err)
	}

	offsets, err = cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if offsets["test1"][0] == 1234 && offsets["test1"][1] == 2345 && offsets["test2"][0] == 3456 {
		t.Log("All offsets present in offset map")
	} else {
		t.Logf("Offset map not as expected: %v", offsets)
	}

	if err := cg.ResetOffsets(); err != nil {
		t.Fatal(err)
	}

	offsets, err = cg.FetchAllOffsets()
	if err != nil {
		t.Error(err)
	}

	if len(offsets) > 0 {
		t.Errorf("After a reset, consumergroup shouldn't have any offsets set, but found offsets for %d topics", len(offsets))
	}
}

func assertWatchTriggered(t *testing.T, ch <-chan zk.Event, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatal("Watch is not triggered")
	}
}

func assertWatchNotTriggered(t *testing.T, ch <-chan zk.Event, timeout time.Duration) {
	select {
	case <-ch:
		t.Fatal("Watch is not supposed to be triggered")
	case <-time.After(timeout):
	}
}
