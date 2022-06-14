package kendra

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusQuerySuggestionsBlockList(ctx context.Context, conn *kendra.Client, id, indexId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
