package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsElasticacheGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticacheGlobalReplicationGroupRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auth_token_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_node_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"global_node_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slots": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"global_replication_group_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatic_failover": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_group_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsElasticacheGlobalReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	groupID := d.Get("global_replication_group_id").(string)

	group, err := getGlobalReplicationGroup(conn, groupID)
	if err != nil {
		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) not found", groupID)
		}
		return fmt.Errorf("reading ElastiCache Global Replication Group (%s): %w", groupID, err)
	}

	if group == nil {
		return fmt.Errorf("reading ElastiCache Global Replication Group (%s): empty output", groupID)
	}

	d.SetId(aws.StringValue(group.GlobalReplicationGroupId))

	d.Set("arn", group.ARN)
	d.Set("at_rest_encryption_enabled", group.AtRestEncryptionEnabled)
	d.Set("auth_token_enabled", group.AuthTokenEnabled)
	d.Set("cache_node_type", group.CacheNodeType)
	d.Set("cluster_enabled", group.ClusterEnabled)
	d.Set("global_replication_group_description", group.GlobalReplicationGroupDescription)
	d.Set("engine", group.Engine)
	d.Set("engine_version", group.EngineVersion)
	d.Set("status", group.Status)
	d.Set("transit_encryption_enabled", group.TransitEncryptionEnabled)

	if group.GlobalNodeGroups != nil {
		var groups []map[string]interface{}
		for _, g := range group.GlobalNodeGroups {
			groups = append(groups, map[string]interface{}{
				"global_node_group_id": aws.StringValue(g.GlobalNodeGroupId),
				"slots":                aws.StringValue(g.Slots),
			})
		}
		d.Set("global_node_groups", groups)
	}

	if group.Members != nil {
		var members []map[string]interface{}
		for _, m := range group.Members {
			members = append(members, map[string]interface{}{
				"automatic_failover":       aws.StringValue(m.AutomaticFailover),
				"replication_group_id":     aws.StringValue(m.ReplicationGroupId),
				"replication_group_region": aws.StringValue(m.ReplicationGroupRegion),
				"role":                     aws.StringValue(m.Role),
				"status":                   aws.StringValue(m.Status),
			})
			if aws.StringValue(m.Role) == "PRIMARY" {
				d.Set("primary_replication_group_id", m.ReplicationGroupId)
			}
		}
		d.Set("members", members)
	}

	return nil
}
