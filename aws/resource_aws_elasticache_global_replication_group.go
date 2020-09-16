package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsElasticacheGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheGlobalReplicationGroupCreate,
		Read:   resourceAwsElasticacheGlobalReplicationGroupRead,
		Update: resourceAwsElasticacheGlobalReplicationGroupUpdate,
		Delete: resourceAwsElasticacheGlobalReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Optional: true,
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
				Optional: true,
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
				Computed: true,
			},
			"global_replication_group_id_suffix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"primary_replication_group_id": {
				Type:     schema.TypeString,
				Required: true,
				Computed: true,
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
		SchemaVersion: 1,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
		},
	}
}

func resourceAwsElasticacheGlobalReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	req := &elasticache.CreateGlobalReplicationGroupInput{}

	if v, ok := d.GetOk("global_replication_group_description"); ok {
		req.GlobalReplicationGroupDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("global_replication_group_id_suffix"); ok {
		req.GlobalReplicationGroupIdSuffix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("primary_replication_group_id"); ok {
		req.PrimaryReplicationGroupId = aws.String(v.(string))
	}

	res, err := conn.CreateGlobalReplicationGroup(req)
	if err != nil {
		return fmt.Errorf("creating ElastiCache Global Replication Group: %w", err)
	}

	d.SetId(aws.StringValue(res.GlobalReplicationGroup.GlobalReplicationGroupId))

	if err := waitForGlobalReplicationGroupCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsElasticacheGlobalReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheGlobalReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	group, err := getGlobalReplicationGroup(conn, d.Id())
	if err != nil {
		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if group == nil {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(group.Status) == "deleting" {
		log.Printf("[WARN] ElastiCache Global Replication Group %s is currently in the `deleting` state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", group.ARN)
	d.Set("at_rest_encryption_enabled", group.AtRestEncryptionEnabled)
	d.Set("auth_token_enabled", group.AuthTokenEnabled)
	d.Set("cache_node_type", group.CacheNodeType)
	d.Set("cluster_enabled", group.ClusterEnabled)
	d.Set("global_replication_group_description", group.GlobalReplicationGroupDescription)
	d.Set("global_replication_group_id", group.GlobalReplicationGroupId)
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

func resourceAwsElasticacheGlobalReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	req := &elasticache.ModifyGlobalReplicationGroupInput{
		// Modifications to Global Replication Groups cannot be requested to
		// be applied in PreferredMaintenceWindow.
		ApplyImmediately:         aws.Bool(true),
		GlobalReplicationGroupId: aws.String(d.Id()),
	}

	modifyRequired := false

	if d.HasChange("automatic_failover_enabled") {
		req.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		modifyRequired = true
	}

	if d.HasChange("cache_node_type") {
		req.CacheNodeType = aws.String(d.Get("cache_node_type").(string))
		modifyRequired = true
	}

	if d.HasChange("global_replication_group_description") {
		req.GlobalReplicationGroupDescription = aws.String(d.Get("global_replication_group_description").(string))
		modifyRequired = true
	}

	if d.HasChange("engine_version") {
		req.EngineVersion = aws.String(d.Get("engine_version").(string))
		modifyRequired = true
	}

	if modifyRequired {
		log.Printf("[DEBUG] Modifying ElastiCache Global Replication Group (%s), opts:\n%s", d.Id(), req)

		if _, err := conn.ModifyGlobalReplicationGroup(req); err != nil {
			return fmt.Errorf("modifying ElastiCache Global Replication Group (%s): %w", d.Id(), err)
		}

		if err := waitForGlobalReplicationGroupModification(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) to finish modifications: %w", d.Id(), err)
		}
	}

	if d.HasChange("primary_replication_group_id") {
		targetPrimaryID := d.Get("primary_replication_group_id").(string)

		var failoverID, failoverRegion string

		members := d.Get("members").([]map[string]interface{})

		for _, m := range members {
			mID := m["replication_group_id"].(string)
			role := m["role"].(string)
			status := m["status"].(string)
			if mID == targetPrimaryID && role == "SECONDARY" && status == "associated" {
				failoverID = mID
				failoverRegion = m["replication_group_region"].(string)
				break
			}
		}

		if failoverID == "" || failoverRegion == "" {
			return fmt.Errorf("new primary replication group %s for ElastiCache Global Replication Group (%s) must already be a member with role SECONDARY and state associated", targetPrimaryID, d.Id())
		}

		failoverReq := &elasticache.FailoverGlobalReplicationGroupInput{
			GlobalReplicationGroupId:  aws.String(d.Id()),
			PrimaryRegion:             aws.String(failoverRegion),
			PrimaryReplicationGroupId: aws.String(failoverID),
		}

		log.Printf("[DEBUG] Failing over ElastiCache Global Replication Group (%s), opts:\n%s", d.Id(), failoverReq)

		if _, err := conn.FailoverGlobalReplicationGroup(failoverReq); err != nil {
			return fmt.Errorf("failing over ElastiCache Global Replication Group (%s): %w", d.Id(), err)
		}

		if err := waitForGlobalReplicationGroupFailover(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for ElastiCache Global Replication Group (%s) to finish failing over: %w", d.Id(), err)
		}
	}

	return resourceAwsElasticacheGlobalReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheGlobalReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	req := &elasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId: aws.String(d.Id()),
		// Global Replication Group and primary replication group are managed
		// as separate resources, so we always want to retain the latter.
		RetainPrimaryReplicationGroup: aws.Bool(true),
	}

	if _, err := conn.DeleteGlobalReplicationGroup(req); err != nil {
		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			return nil
		}
		return fmt.Errorf("deleting Global Replication Group (%s): %w", d.Id(), err)
	}

	if err := waitForGlobalReplicationGroupDeletion(conn, d.Id(), 40*time.Minute); err != nil {
		return fmt.Errorf("waiting for Global Replication Group (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func getGlobalReplicationGroup(conn *elasticache.ElastiCache, id string) (*elasticache.GlobalReplicationGroup, error) {
	req := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(id),
		ShowMemberInfo:           aws.Bool(true),
	}

	res, err := conn.DescribeGlobalReplicationGroups(req)
	if err != nil {
		return nil, err
	}

	if len(res.GlobalReplicationGroups) == 0 {
		return nil, nil
	}

	group := res.GlobalReplicationGroups[0]
	if aws.StringValue(group.GlobalReplicationGroupId) != id {
		return nil, nil
	}

	return group, nil
}

func globalReplicationGroupStateRefreshFunc(conn *elasticache.ElastiCache, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		group, err := getGlobalReplicationGroup(conn, id)
		if err != nil {
			return nil, "", err
		}
		if group == nil {
			return nil, "", nil
		}

		status := aws.StringValue(group.Status)
		log.Printf("[INFO] ElastiCache Global Replication Group (%s) status: %s", id, status)
		return group, status, nil
	}
}

func waitForGlobalReplicationGroupCreation(conn *elasticache.ElastiCache, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available", "primary-only"},
		Refresh:    globalReplicationGroupStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache  Global Replication Group (%s) creation", id)

	_, err := stateConf.WaitForState()
	return err
}

func waitForGlobalReplicationGroupModification(conn *elasticache.ElastiCache, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"available", "primary-only"},
		Refresh:    globalReplicationGroupStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) modification", id)

	_, err := stateConf.WaitForState()
	return err
}

func waitForGlobalReplicationGroupFailover(conn *elasticache.ElastiCache, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"available"},
		Refresh:    globalReplicationGroupStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) failover", id)

	_, err := stateConf.WaitForState()
	return err
}

func waitForGlobalReplicationGroupDeletion(conn *elasticache.ElastiCache, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"available", "creating", "deleting", "incompatible-network", "incompatible-parameters", "primary-only", "restore-failed", "snapshotting"},
		Target:     []string{},
		Refresh:    globalReplicationGroupStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for ElastiCache Global Replication Group (%s) deletion", id)

	_, err := stateConf.WaitForState()
	return err
}
