package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/verify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketCreate,
		Read:   resourceBucketRead,
		Update: resourceBucketUpdate,
		Delete: resourceBucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"versioning_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     TagsSchema(),
			"tags_all": TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateBucketInput{
		BucketName: aws.String(d.Get("name").(string)),
		BundleId:   aws.String(d.Get("bundle_id").(string)),
	}

	if v, ok := d.GetOk("versioning_enabled"); ok {
		req.EnableObjectVersioning = aws.Bool(v.(bool))
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateBucket(context.TODO(), &req)

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Bucket (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(d.Get("name").(string))

	return resourceBucketRead(d, meta)
}

func resourceBucketRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetBuckets(context.TODO(), &lightsail.GetBucketsInput{
		BucketName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	b := resp.Buckets[0]

	d.Set("arn", b.Arn)
	d.Set("name", b.Name)
	d.Set("bundle_id", b.BundleId)
	d.Set("url", b.Url)
	d.Set("created_at", b.CreatedAt.Format(time.RFC3339))
	d.Set("support_code", b.SupportCode)

	if aws.ToString(b.ObjectVersioning) == "Enabled" {
		d.Set("versioning_enabled", true)
	} else {
		d.Set("versioning_enabled", false)
	}

	tags := KeyValueTags(b.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceBucketUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("versioning_enabled") {
		req := lightsail.UpdateBucketInput{
			BucketName: aws.String(d.Id()),
		}
		if d.Get("versioning_enabled").(bool) {
			req.Versioning = aws.String("Enabled")
		} else {
			req.Versioning = aws.String("Suspended")
		}
		resp, err := conn.UpdateBucket(context.TODO(), &req)
		op := resp.Operations[0]

		err = waitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Bucket (%s) to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("bundle_id") {
		req := lightsail.UpdateBucketBundleInput{
			BucketName: aws.String(d.Id()),
			BundleId:   aws.String(d.Get("bundle_id").(string)),
		}
		resp, err := conn.UpdateBucketBundle(context.TODO(), &req)
		op := resp.Operations[0]

		err = waitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Bucket (%s) to become ready: %s", d.Id(), err)
		}
	}

	return resourceBucketRead(d, meta)
}

func resourceBucketDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	_, err := conn.DeleteBucket(context.TODO(), &lightsail.DeleteBucketInput{
		BucketName: aws.String(d.Id()),
	})

	return err
}
