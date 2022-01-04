package lightsail

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TagsSchema returns the schema to use for tags.
//
func TagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

func TagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

// []*SERVICE.Tag handling

// Tags returns lightsail service tags.
func Tags(tags tftags.KeyValueTags) []types.Tag {
	result := make([]types.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from lightsail service tags.
func KeyValueTags(tags []types.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates lightsail service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *lightsail.Client, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &lightsail.UntagResourceInput{
			ResourceName: aws.String(identifier),
			TagKeys:      removedTags.IgnoreAWS().Keys(),
		}

		_, err := conn.UntagResource(context.TODO(), input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &lightsail.TagResourceInput{
			ResourceName: aws.String(identifier),
			Tags:         Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagResource(context.TODO(), input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// // []*SERVICE.Tag handling

// // Tags returns lightsail service tags.
// func Tags(tags interface{}) []types.Tag {
// 	result := []types.Tag{}

// 	for k, v := range tags.(map[string]interface{}) {
// 		tag := types.Tag{
// 			Key:   aws.String(k),
// 			Value: aws.String(fmt.Sprintf("%v", v)),
// 		}

// 		result = append(result, tag)
// 	}

// 	return result
// }

// // TagsToMap returns lightsail service tags as a Map
// func TagsToMap(tags []types.Tag) map[string]string {
// 	result := make(map[string]string)

// 	for _, v := range tags {
// 		result[*v.Key] = *v.Value
// 	}

// 	return result
// }

// // UpdateTags updates lightsail service tags.
// // The identifier is typically the Amazon Resource Name (ARN), although
// // it may also be a different identifier depending on the service.
// func UpdateTags(conn *lightsail.Client, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
// 	oldTags := oldTagsMap.(map[string]interface{})
// 	newTags := newTagsMap.(map[string]interface{})

// 	removedTags := []string{}
// 	for k := range oldTags {
// 		if _, ok := newTags[k]; !ok {
// 			removedTags = append(removedTags, fmt.Sprintf("%v", k))
// 		}
// 	}

// 	if len(removedTags) > 0 {
// 		input := &lightsail.UntagResourceInput{
// 			ResourceName: aws.String(identifier),
// 			TagKeys:      removedTags,
// 		}

// 		_, err := conn.UntagResource(context.TODO(), input)

// 		if err != nil {
// 			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
// 		}
// 	}

// 	if len(newTags) > 0 {
// 		input := &types.TagResourceInput{
// 			ResourceName: aws.String(identifier),
// 			Tags:         Tags(newTags),
// 		}

// 		_, err := conn.TagResource(context.TODO(), input)

// 		if err != nil {
// 			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
// 		}
// 	}

// 	return nil
// }
