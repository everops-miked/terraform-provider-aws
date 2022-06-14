package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceQuerySuggestionsBlockList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQuerySuggestionsBlockListCreate,
		ReadWithoutTimeout:   resourceQuerySuggestionsBlockListRead,
		UpdateWithoutTimeout: resourceQuerySuggestionsBlockListUpdate,
		DeleteWithoutTimeout: resourceQuerySuggestionsBlockListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source_s3_path": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceQuerySuggestionsBlockListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	in := &kendra.CreateQuerySuggestionsBlockListInput{
		ClientToken:  aws.String(resource.UniqueId()),
		IndexId:      aws.String(d.Get("index_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		SourceS3Path: expandSourceS3Path(d.Get("source_s3_path").([]interface{})),
	}

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateQuerySuggestionsBlockList(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Kendra QuerySuggestionsBlockList (%s): %s", d.Get("name").(string), err)
	}

	if out == nil {
		return diag.Errorf("creating Amazon Kendra QuerySuggestionsBlockList (%s): empty output", d.Get("name").(string))
	}

	id := aws.ToString(out.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitQuerySuggestionsBlockListCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Kendra QuerySuggestionsBlockList (%s) create: %s", d.Id(), err)
	}

	return resourceQuerySuggestionsBlockListRead(ctx, d, meta)
}

func resourceQuerySuggestionsBlockListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	out, err := FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra QuerySuggestionsBlockList (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "kendra",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("querysuggestionsblocklist/%s", id),
	}.String()

	d.Set("arn", arn)
	d.Set("description", out.Description)
	d.Set("index_id", out.IndexId)
	d.Set("name", out.Name)
	d.Set("role_arn", out.RoleArn)
	d.Set("status", out.Status)

	if err := d.Set("source_s3_path", flattenSourceS3Path(out.SourceS3Path)); err != nil {
		return diag.Errorf("setting complex argument: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("listing tags for Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceQuerySuggestionsBlockListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	if d.HasChangesExcept("tags", "tags_all") {
		id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		input := &kendra.UpdateQuerySuggestionsBlockListInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("source_s3_path") {
			input.SourceS3Path = expandSourceS3Path(d.Get("source_s3_path").([]interface{}))
		}

		log.Printf("[DEBUG] Updating Kendra QuerySuggestionsBlockList (%s): %#v", d.Id(), input)
		_, err = conn.UpdateQuerySuggestionsBlockList(ctx, input)
		if err != nil {
			return diag.Errorf("updating Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
		}

		if _, err := waitQuerySuggestionsBlockListUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Kendra QuerySuggestionsBlockList (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Kendra QuerySuggestionsBlockList (%s) tags: %s", d.Id(), err))
		}
	}

	return resourceQuerySuggestionsBlockListRead(ctx, d, meta)
}

func resourceQuerySuggestionsBlockListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	log.Printf("[INFO] Deleting Kendra QuerySuggestionsBlockList %s", d.Id())

	id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteQuerySuggestionsBlockList(ctx, &kendra.DeleteQuerySuggestionsBlockListInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var notFound *types.ResourceNotFoundException

	if errors.As(err, &notFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
	}

	if _, err := waitQuerySuggestionsBlockListDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra QuerySuggestionsBlockList (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func flattenSourceS3Path(apiObject *types.S3Path) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		m["bucket"] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		m["key"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandSourceS3Path(tfList []interface{}) *types.S3Path {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Path{}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}
