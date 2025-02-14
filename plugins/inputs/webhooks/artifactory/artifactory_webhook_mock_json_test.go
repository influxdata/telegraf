package artifactory

func unsupportedEventJSON() string {
	return `
	{
		"domain": "not_supported",
		"event_type": "not_supported",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0
		}
	}`
}

func artifactDeployedEventJSON() string {
	return `
	{
		"domain": "artifact",
		"event_type": "deployed",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0
		}
	}`
}

func artifactDeletedEventJSON() string {
	return `
	{
		"domain": "artifact",
		"event_type": "deleted",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0
		}
	}`
}

func artifactMovedEventJSON() string {
	return `
	{
		"domain": "artifact",
		"event_type": "moved",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0,
		  "source_repo_path": "sample_repo",
		  "target_repo_path": "sample_target_repo"
		}
	}`
}

func artifactCopiedEventJSON() string {
	return `
	{
		"domain": "artifact",
		"event_type": "copied",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0,
		  "source_repo_path": "sample_repo",
		  "target_repo_path": "sample_target_repo"
		}
	}`
}

func artifactPropertiesAddedEventJSON() string {
	return `
	{
		"domain": "artifact_property",
		"event_type": "added",
		"data": {
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "property_key": "sample_key",
		  "property_values": [
			"sample_value1"
		  ],
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0
		}
	}`
}

func artifactPropertiesDeletedEventJSON() string {
	return `
	{
		"domain": "artifact_property",
		"event_type": "deleted",
		"data": {
			"name": "sample.txt",
			"path": "sample_dir/sample.txt",
			"property_key": "sample_key",
			"property_values": [
			  "sample_value1"
			],
			"repo_key": "sample_repo",
			"sha256": "sample_checksum",
			"size": 0
		}
	}`
}

func dockerPushedEventJSON() string {
	return `
	{
		"domain": "docker",
		"event_type": "pushed",
		"data": {
		  "image_name": "sample_arch",
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "platforms": [
			{
			  "architecture": "sample_os",
			  "os": "sample_tag"
			}
		  ],
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0,
		  "tag": "sample_image"
		}
	}`
}

func dockerDeletedEventJSON() string {
	return `
	{
		"domain": "docker",
		"event_type": "deleted",
		"data": {
		  "image_name": "sample_arch",
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "platforms": [
			{
			  "architecture": "sample_os",
			  "os": "sample_tag"
			}
		  ],
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0,
		  "tag": "sample_image"
		}
	}`
}

func dockerPromotedEventJSON() string {
	return `
	{
		"domain": "docker",
		"event_type": "promoted",
		"data": {
		  "image_name": "sample_arch",
		  "name": "sample.txt",
		  "path": "sample_dir/sample.txt",
		  "platforms": [
			{
			  "architecture": "sample_os",
			  "os": "sample_tag"
			}
		  ],
		  "repo_key": "sample_repo",
		  "sha256": "sample_checksum",
		  "size": 0,
		  "tag": "sample_image"
		}
	}`
}

func buildUploadedEventJSON() string {
	return `
	{
		"domain": "build",
		"event_type": "uploaded",
		"data": {
		  "build_name": "sample_build_name",
		  "build_number": "1",
		  "build_started": "1970-01-01T00:00:00.000+0000"
		}
	}`
}

func buildDeletedEventJSON() string {
	return `
	{
		"domain": "build",
		"event_type": "deleted",
		"data": {
		  "build_name": "sample_build_name",
		  "build_number": "1",
		  "build_started": "1970-01-01T00:00:00.000+0000"
		}
	}`
}

func buildPromotedEventJSON() string {
	return `
	{
		"domain": "build",
		"event_type": "promoted",
		"data": {
		  "build_name": "sample_build_name",
		  "build_number": "1",
		  "build_started": "1970-01-01T00:00:00.000+0000"
		}
	}`
}

func releaseBundleCreatedEventJSON() string {
	return `
	{
		"domain": "release_bundle",
		"event_type": "created",
		"destination": "release_bundle",
		"data": {
			"release_bundle_name": "sample_name",
			"release_bundle_size": 9800,
			"release_bundle_version": "1.0.0"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func releaseBundleSignedEventJSON() string {
	return `
	{
		"domain": "release_bundle",
		"event_type": "signed",
		"destination": "release_bundle",
		"data": {
			"release_bundle_name": "sample_name",
			"release_bundle_size": 9800,
			"release_bundle_version": "1.0.0"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func releaseBundleDeletedEventJSON() string {
	return `
	{
		"domain": "release_bundle",
		"event_type": "signed",
		"destination": "release_bundle",
		"data": {
			"release_bundle_name": "sample_name",
			"release_bundle_size": 9800,
			"release_bundle_version": "1.0.0"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func distributionStartedEventJSON() string {
	return `
	{
		"domain": "distribution",
		"event_type": "distribute_started",
		"destination": "distribution",
		"data": {
		  "edge_node_info_list": [
			{
			  "edge_node_address": "https://artifactory-edge2-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge2"
			},
			{
			  "edge_node_address": "https://artifactory-edge1-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge1"
			}
		  ],
		  "release_bundle_name": "test",
		  "release_bundle_size": 1037976,
		  "release_bundle_version": "1.0.0",
		  "status_message": "CREATED",
		  "transaction_id": 395969746957422600
		},
		"jpd_origin": "https://ga-dev.jfrogdev.co/artifactory"
	  }`
}

func distributionCompletedEventJSON() string {
	return `
	{
		"domain": "distribution",
		"event_type": "distribute_completed",
		"destination": "distribution",
		"data": {
		  "edge_node_info_list": [
			{
			  "edge_node_address": "https://artifactory-edge2-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge2"
			},
			{
			  "edge_node_address": "https://artifactory-edge1-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge1"
			}
		  ],
		  "release_bundle_name": "test",
		  "release_bundle_size": 1037976,
		  "release_bundle_version": "1.0.0",
		  "status_message": "CREATED",
		  "transaction_id": 395969746957422600
		},
		"jpd_origin": "https://ga-dev.jfrogdev.co/artifactory"
	  }`
}

func distributionAbortedEventJSON() string {
	return `
	{
		"domain": "distribution",
		"event_type": "distribute_aborted",
		"destination": "distribution",
		"data": {
		  "edge_node_info_list": [
			{
			  "edge_node_address": "https://artifactory-edge2-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge2"
			},
			{
			  "edge_node_address": "https://artifactory-edge1-dev.jfrogdev.co/artifactory",
			  "edge_node_name": "artifactory-edge1"
			}
		  ],
		  "release_bundle_name": "test",
		  "release_bundle_size": 1037976,
		  "release_bundle_version": "1.0.0",
		  "status_message": "CREATED",
		  "transaction_id": 395969746957422600
		},
		"jpd_origin": "https://ga-dev.jfrogdev.co/artifactory"
	  }`
}

func distributionFailedEventJSON() string {
	return `
	{
		"domain": "distribution",
		"event_type": "distribute_failed",
		"destination": "distribution",
		"data": {
			"edge_node_info_list": [
		  	{
				"edge_node_address": "https://artifactory-edge2-dev.jfrogdev.co/artifactory",
				"edge_node_name": "artifactory-edge2"
		  	},
			{
				"edge_node_address": "https://artifactory-edge1-dev.jfrogdev.co/artifactory",
				"edge_node_name": "artifactory-edge1"
		  	}
			],
			"release_bundle_name": "test",
			"release_bundle_size": 1037976,
			"release_bundle_version": "1.0.0",
			"status_message": "CREATED",
			"transaction_id": 395969746957422600
	  	},
	  "jpd_origin": "https://ga-dev.jfrogdev.co/artifactory"
	}`
}

func destinationReceivedEventJSON() string {
	return `
	{
		"domain": "destination",
		"event_type": "received",
		"destination": "artifactory_release_bundle",
		"data": {
			"release_bundle_name": "test",
			"release_bundle_version": "1.0.0",
			"status_message": "COMPLETED"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func destinationDeleteStartedEventJSON() string {
	return `
	{
		"domain": "destination",
		"event_type": "delete_started",
		"destination": "artifactory_release_bundle",
	  	"data": {
			"release_bundle_name": "test",
			"release_bundle_version": "1.0.0",
			"status_message": "COMPLETED"
	  	},
	  	"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func destinationDeleteCompletedEventJSON() string {
	return `
	{
		"domain": "destination",
		"event_type": "delete_completed",
		"destination": "artifactory_release_bundle",
		"data": {
		  "release_bundle_name": "test",
		  "release_bundle_version": "1.0.0",
		  "status_message": "COMPLETED"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}

func destinationDeleteFailedEventJSON() string {
	return `
	{
		"domain": "destination",
		"event_type": "delete_failed",
		"destination": "artifactory_release_bundle",
		"data": {
		  "release_bundle_name": "test",
		  "release_bundle_version": "1.0.0",
		  "status_message": "COMPLETED"
		},
		"jpd_origin": "https://dist-pipe2.jfrogdev.co/artifactory"
	}`
}
