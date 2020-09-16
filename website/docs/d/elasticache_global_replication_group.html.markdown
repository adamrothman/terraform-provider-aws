---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_global_replication_group"
description: |-
  Get information on an ElastiCache Global Replication Group resource.
---

# Data Source: aws_elasticache_global_replication_group

Use this data source to get information about an ElastiCache Global Replication Group.

## Example Usage

```hcl
data "aws_elasticache_global_replication_group" "bar" {
  global_replication_group_id = "sgaui-my-global-datastore"
}
```

## Argument Reference

The following arguments are supported:

* `global_replication_group_id` – (Required) The identifier for the global replication group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN (Amazon Resource Name) of the global replication group. Will be of format `arn:aws:elasticache::{account-id}:globalreplicationgroup:{identifier}`.
* `at_rest_encryption_enabled` - A flag that enables encryption at rest when set to true. You cannot modify the value of AtRestEncryptionEnabled after the replication group is created. To enable encryption at rest on a replication group you must set AtRestEncryptionEnabled to true when you create the replication group.
* `auth_token_enabled` - A flag that enables using an AuthToken (password) when issuing Redis commands.
* `cache_node_type` - The cache node type of the Global Datastore.
* `cluster_enabled` - A flag that indicates whether the Global Datastore is cluster enabled.
* `engine` - The Elasticache engine. For Redis only.
* `engine_version` - The Elasticache Redis engine version. For preview, it is Redis version 5.0.5 only.
* `global_node_groups` - Indicates the slot configuration and global identifier for each slice group.
* `global_replication_group_description` - The optional description of the Global Datastore.
* `global_replication_group_id` - The name of the Global Datastore.
* `members` - The replication groups that comprise the Global Datastore.
* `status` - The status of the Global Datastore.
* `transit_encryption_enabled` - A flag that enables in-transit encryption when set to true. You cannot modify the value of TransitEncryptionEnabled after the cluster is created. To enable in-transit encryption on a cluster you must set TransitEncryptionEnabled to true when you create a cluster.
