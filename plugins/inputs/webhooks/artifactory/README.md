# artifactory webhook

You need to configure to orginizations artifactory instance/s as detailed via the artifactory webhook documentation: <https://www.jfrog.com/confluence/display/JFROG/Webhooks>.  Multiple webhooks may need be needed to configure different domains.

You can also add a secret that will be used by telegraf to verify the authenticity of the requests.

## Events

The different events type can be found found in the webhook documentation: <https://www.jfrog.com/confluence/display/JFROG/Webhooks>.  Events are identified by their `domain` and `event`.  The following sections break down each event by domain.

### Artifact Domain

#### Artifact Deployed Event

The Webhook is triggered when an artifact is deployed to a repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string

**Fields:**

* 'size' int
* 'sha256' string

#### Artifact Deleted Event

The Webhook is triggered when an artifact is deleted from a repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string

**Fields:**

* 'size' int
* 'sha256' string

#### Artifact Moved Event

The Webhook is triggered when an artifact is moved from a repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string

**Fields:**

* 'size' int
* 'source_path' string
* 'target_path' string

#### Artifact Copied Event

The Webhook is triggered when an artifact is copied from a repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string

**Fields:**

* 'size' int
* 'source_path' string
* 'target_path' string

### Artifact Properties Domain

#### Properties Added Event

The Webhook is triggered when a property is added to an artifact/folder in a repository, or the repository itself.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
**Fields**
* 'property_key' string
* 'property_values' string (joined comma seperated list)

#### Properties Deleted Event

The Webhook is triggered when a property is deleted from an artifact/folder in a repository, or the repository itself.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string

**Fields:**

* 'property_key' string
* 'property_values' string (joined comma seperated list)

### Docker Domain

#### Docker Pushed Event

The Webhook is triggered when a new tag of a Docker image is pushed to a Docker repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string

**Fields:**

* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
  * 'achitecture' string
  * 'os' string

#### Docker Deleted Event

The Webhook is triggered when a tag of a Docker image is deleted from a Docker repository.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string

**Fields:**

* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
  * 'achitecture' string
  * 'os' string

#### Docker Promoted Event

The Webhook is triggered when a tag of a Docker image is promoted.

**Tags:**

* 'domain' string
* 'event_type' string
* 'repo' string
* 'path' string
* 'name' string
* 'image_name' string

**Fields:**

* 'size' string
* 'sha256' string
* 'tag' string
* 'platforms' []object
  * 'achitecture' string
  * 'os' string

### Build Domain

#### Build Uploaded Event

The Webhook is triggered when a new build is uploaded.

**Tags:**

* 'domain' string
* 'event_type' string

**Fields:**

* 'build_name' string
* 'build_number' string
* 'build_started' string

#### Build Deleted Event

The Webhook is triggered when a build is deleted.

**Tags:**

* 'domain' string
* 'event_type' string

**Fields:**

* 'build_name' string
* 'build_number' string
* 'build_started' string

#### Build Promoted Event

The Webhook is triggered when a build is promoted.

**Tags:**

* 'domain' string
* 'event_type' string

**Fields:**

* 'build_name' string
* 'build_number' string
* 'build_started' string

### Release Bundle Domain

#### Release Bundle Created Event

The Webhook is triggered when a Release Bundle is created.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string

**Fields:**

* 'release_bundle_name' string
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

#### Release Bundle Signed Event

The Webhook is triggered when a Release Bundle is signed.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string

**Fields:**

* 'release_bundle_name' string
* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

#### Release Bundle Deleted Event

The Webhook is triggered when a Release Bundle is deleted.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'jpd_origin' string

### Release Bundle Distribution Domain

#### Release Bundle Distribution Started Event

The Webhook is triggered when Release Bundle distribution has started

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

#### Release Bundle Distribution Completed Event

The Webhook is triggered when Release Bundle distribution has completed.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

#### Release Bundle Distribution Aborted Event

The Webhook is triggered when Release Bundle distribution has been aborted.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

#### Release Bundle Distribution Failed Event

The Webhook is triggered when Release Bundle distribution has failed.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

### Release Bundle Version Domain

#### Release Bundle Version Deletion Started EVent

The Webhook is triggered when a Release Bundle version deletion has started on one or more Edge nodes.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

#### Release Bundle Version Deletion Completed Event

The Webhook is triggered when a Release Bundle version deletion has completed from one or more Edge nodes.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

#### Release Bundle Version Deletion Failed Event

The Webhook is triggered when a Release Bundle version deletion has failed on one or more Edge nodes.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_size' string
* 'release_bundle_version' string
* 'status_message' string
* 'transaction_id' string
* 'edge_node_info_list' []object
  * 'edge_node_address' string
  * 'edge_node_name' string
* 'jpd_origin' string

### Release Bundle Destination Domain

#### Release Bundle Received Event

The Webhook is triggered when a Release Bundle was received on an Edge Node.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

### Release Bundle Delete Started Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node completed.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

#### Release Bundle Delete Completed Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node completed.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string

#### Release Bundle Delete Failed Event

The Webhook is triggered when a Release Bundle deletion from an Edge Node fails.

**Tags:**

* 'domain' string
* 'event_type' string
* 'destination' string
* 'release_bundle_name' string

**Fields:**

* 'release_bundle_version' string
* 'status_message' string
* 'jpd_origin' string
