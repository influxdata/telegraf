package kazoo

import (
	"testing"
	"time"

	//	"github.com/samuel/go-zookeeper/zk"
	"github.com/samuel/go-zookeeper/zk"
	"reflect"
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
	if reg, err := instance1.Registration(); err != zk.ErrNoNode || reg != nil {
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
	if err := instance.Register([]string{"topic"}); err != nil {
		t.Fatal(err)
	}

	// The instance watch should have been triggered
	<-c

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
	<-c

	instances, err = cg.Instances()
	if err != nil {
		t.Fatal(err)
	}

	if len(instances) != 0 {
		t.Error("Expected 0 running instances")
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

	if offset != 0 {
		t.Error("Expected to get offset 0 for a partition that hasn't seen an offset commit yet")
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
