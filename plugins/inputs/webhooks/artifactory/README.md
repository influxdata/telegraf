# artifactory webhook

You need to configure to orginizations artifactory instance/s as detailed via the artifactory [webhook documentation] (https://www.jfrog.com/confluence/display/JFROG/Webhooks).  Multiple webhooks may need be needed to configure different domains.

You can also add a secret that will be used by telegraf to verify the authenticity of the requests.

## Events

The different events type can be found found in the [webhook documentation] (https://www.jfrog.com/confluence/display/JFROG/Webhooks).  Events are identified by their `domain` and `event`.  The following sections break down each event by domain.

### Artifact Domain

#### Deployed Event

The Webhook is triggered when an artifact is deployed to a repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'size' int
* 'sha256' string

#### Deleted Event

The Webhook is triggered when an artifact is deleted from a repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'size' int
* 'sha256' string

#### Moved Event

The Webhook is triggered when an artifact is moved from a repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'size' int
* 'source_path' string
* 'target_path' string

#### Copied Event

The Webhook is triggered when an artifact is copied from a repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'size' int
* 'source_path' string
* 'target_path' string

### Artifact Properties Domain

#### Added Event

The Webhook is triggered when a property is added to an artifact/folder in a repository, or the repository itself.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'property_key' string
* 'property_values' string (joined comma seperated list)

#### Deleted Event

The Webhook is triggered when a property is deleted from an artifact/folder in a repository, or the repository itself.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'property_key' string
* 'property_values' string (joined comma seperated list)

### Docker Domain

#### Pushed Event

The Webhook is triggered when a new tag of a Docker image is pushed to a Docker repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string
**Fields**
* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
    * 'achitecture' string
    * 'os' string

#### Deleted Event

The Webhook is triggered when a tag of a Docker image is deleted from a Docker repository.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string
**Fields**
* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
    * 'achitecture' string
    * 'os' string

#### Promoted Event

The Webhook is triggered when a tag of a Docker image is promoted.

**Tags**
* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string
**Fields**
* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
    * 'achitecture' string
    * 'os' string

### Build Domain

#### Uploaded Event

The Webhook is triggered when a new build is uploaded.

**Tags**
* 'domain' string
* 'event_type' string
**Fields**
* 'build_name' string
* 'build_number' string
* 'build_started' string

#### Deleted Event

The Webhook is triggered when a build is deleted.

**Tags**
* 'domain' string
* 'event_type' string
**Fields**
* 'build_name' string
* 'build_number' string
* 'build_started' string

#### Promoted Event

The Webhook is triggered when a build is promoted.

**Tags**
* 'domain' string
* 'event_type' string
**Fields**
* 'build_name' string
* 'build_number' string
* 'build_started' string

### Release Bundle Domain

#### Created Event

The Webhook is triggered when a Release Bundle is created.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
**Fields**
* 'release_bundle_name' string
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

#### Signed Event

The Webhook is triggered when a Release Bundle is signed.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
**Fields**
* 'release_bundle_name' string
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

#### Deleted Event

The Webhook is triggered when a Release Bundle is deleted.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

### Distribution Domain

#### Distibute Started Event

The Webhook is triggered when Release Bundle distribution has started

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Distribute Completed Event

The Webhook is triggered when Release Bundle distribution has completed.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Distribute Aborted Event

The Webhook is triggered when Release Bundle distribution has been aborted.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Distribute Failed Event

The Webhook is triggered when Release Bundle distribution has failed.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Deletion Started EVent

The Webhook is triggered when a Release Bundle version deletion has started on one or more Edge nodes.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Deletion Completed Event

The Webhook is triggered when a Release Bundle version deletion has completed from one or more Edge nodes.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

#### Deletion Failed Event

The Webhook is triggered when a Release Bundle version deletion has failed on one or more Edge nodes.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
    * 'edge_node_address' string
    * 'edge_node_name' string
* 'jpd_origin' string

### Destination Domain

#### Received Event

The Webhook is triggered when a Release Bundle was received on an Edge Node.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

### Delete Started Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node completed.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

#### Delete Completed Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node completed. 

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

#### Delete Failed Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node fails.

**Tags**
* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string
**Fields**
* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string
