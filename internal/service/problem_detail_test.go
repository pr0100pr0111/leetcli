package service

import (
	"context"
	"testing"
)

const twoSumDetailResponse = `{
	"data": {
		"question": {
			"questionId": "1",
			"questionFrontendId": "1",
			"title": "Two Sum",
			"titleSlug": "two-sum",
			"difficulty": "Easy",
			"sampleTestCase": "[2,7,11,15]\n9",
			"metaData": "{\"name\":\"twoSum\",\"params\":[{\"name\":\"nums\",\"type\":\"integer[]\"},{\"name\":\"target\",\"type\":\"integer\"}],\"return\":{\"type\":\"integer[]\"}}",
			"content": "<p>You are given an array of integers <code>nums</code>&nbsp;and an integer <code>target</code>.</p>\n\n<strong class=\"example\">Example 1:</strong>\n\n<pre>\n<strong>Input:</strong> nums = [2,7,11,15], target = 9\n<strong>Output:</strong> [0,1]\n<strong>Explanation:</strong> Because nums[0] + nums[1] == 9, we return [0, 1].\n</pre>\n\n<strong class=\"example\">Example 2:</strong>\n\n<pre>\n<strong>Input:</strong> nums = [3,2,4], target = 6\n<strong>Output:</strong> [1,2]\n</pre>",
			"codeSnippets": [
				{"lang": "Go", "langSlug": "golang", "code": "func twoSum(nums []int, target int) []int {\n    \n}"},
				{"lang": "Python3", "langSlug": "python3", "code": "class Solution:\n    def twoSum(self, nums: list[int], target: int) -> list[int]:\n        "},
				{"lang": "Rust", "langSlug": "rust", "code": "impl Solution {\n    pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32> {\n        \n    }\n}"}
			]
		}
	}
}`

const linkedListDetailResponse = `{
	"data": {
		"question": {
			"questionId": "2",
			"questionFrontendId": "2",
			"title": "Add Two Numbers",
			"titleSlug": "add-two-numbers",
			"difficulty": "Medium",
			"sampleTestCase": "[2,4,3]\n[5,6,4]",
			"metaData": "{\r\n  \"name\": \"addTwoNumbers\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"l1\",\r\n      \"type\": \"ListNode\",\r\n      \"dealloc\": false\r\n    },\r\n    {\r\n      \"name\": \"l2\",\r\n      \"type\": \"ListNode\",\r\n      \"dealloc\": false\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"ListNode\",\r\n    \"dealloc\": true\r\n  }\r\n}",
			"content": "<p>Add two numbers represented by linked lists.</p>\n<strong class=\"example\">Example 1:</strong>\n<pre>\n<strong>Input:</strong> l1 = [2,4,3], l2 = [5,6,4]\n<strong>Output:</strong> [7,0,8]\n<strong>Explanation:</strong> 342 + 465 = 807.\n</pre>",
			"codeSnippets": [
				{"lang": "Go", "langSlug": "golang", "code": "func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {\n    \n}"}
			]
		}
	}
}`

func TestGetProblemDetail(t *testing.T) {
	client := &mockClient{detailResponse: []byte(twoSumDetailResponse)}
	service := NewProfileService(client, "testuser")

	d, err := service.GetProblemDetail(context.Background(), "two-sum")
	if err != nil {
		t.Fatalf("GetProblemDetail returned error: %v", err)
	}

	if client.detailSlug != "two-sum" {
		t.Errorf("slug passed to client = %q, want %q", client.detailSlug, "two-sum")
	}
	if d.ID != 1 || d.Title != "Two Sum" || d.Difficulty != "Easy" {
		t.Errorf("detail = %+v", d)
	}
	if d.SampleInput != "[2,7,11,15]\n9" {
		t.Errorf("SampleInput = %q", d.SampleInput)
	}

	if d.Signature.Name != "twoSum" {
		t.Errorf("Signature.Name = %q, want %q", d.Signature.Name, "twoSum")
	}
	if len(d.Signature.Params) != 2 {
		t.Fatalf("params count = %d, want 2", len(d.Signature.Params))
	}
	if d.Signature.Params[0].Name != "nums" || d.Signature.Params[0].Type != "integer[]" {
		t.Errorf("param[0] = %+v", d.Signature.Params[0])
	}
	if d.Signature.Params[1].Name != "target" || d.Signature.Params[1].Type != "integer" {
		t.Errorf("param[1] = %+v", d.Signature.Params[1])
	}
	if d.Signature.Return != "integer[]" {
		t.Errorf("Return = %q, want %q", d.Signature.Return, "integer[]")
	}

	if len(d.Snippets) != 3 {
		t.Fatalf("snippets count = %d, want 3", len(d.Snippets))
	}
	sn, ok := d.Snippet("golang")
	if !ok {
		t.Fatal("golang snippet not found")
	}
	if sn.Lang != "Go" {
		t.Errorf("snippet lang = %q, want %q", sn.Lang, "Go")
	}
	if sn.Code != "func twoSum(nums []int, target int) []int {\n    \n}" {
		t.Errorf("golang snippet code = %q", sn.Code)
	}
	if _, ok := d.Snippet("nosuchlang"); ok {
		t.Errorf("Snippet(nosuchlang) should return false")
	}

	if len(d.Examples) != 2 {
		t.Fatalf("examples count = %d, want 2", len(d.Examples))
	}
	ex1 := d.Examples[0]
	if ex1.Number != 1 {
		t.Errorf("example 1 number = %d", ex1.Number)
	}
	if ex1.Input != "nums = [2,7,11,15], target = 9" {
		t.Errorf("example 1 input = %q", ex1.Input)
	}
	if ex1.Output != "[0,1]" {
		t.Errorf("example 1 output = %q", ex1.Output)
	}
	if ex1.Explanation != "Because nums[0] + nums[1] == 9, we return [0, 1]." {
		t.Errorf("example 1 explanation = %q", ex1.Explanation)
	}
	ex2 := d.Examples[1]
	if ex2.Number != 2 || ex2.Input != "nums = [3,2,4], target = 6" || ex2.Output != "[1,2]" {
		t.Errorf("example 2 = %+v", ex2)
	}
	if ex2.Explanation != "" {
		t.Errorf("example 2 explanation = %q, want empty", ex2.Explanation)
	}
}

func TestGetProblemDetail_ListNodeSignature(t *testing.T) {
	client := &mockClient{detailResponse: []byte(linkedListDetailResponse)}
	service := NewProfileService(client, "testuser")

	d, err := service.GetProblemDetail(context.Background(), "add-two-numbers")
	if err != nil {
		t.Fatalf("GetProblemDetail returned error: %v", err)
	}

	if d.Signature.Name != "addTwoNumbers" {
		t.Errorf("Signature.Name = %q", d.Signature.Name)
	}
	if len(d.Signature.Params) != 2 || d.Signature.Params[0].Type != "ListNode" {
		t.Errorf("params = %+v", d.Signature.Params)
	}
	if d.Signature.Return != "ListNode" {
		t.Errorf("Return = %q", d.Signature.Return)
	}
	if d.SampleInput != "[2,4,3]\n[5,6,4]" {
		t.Errorf("SampleInput = %q", d.SampleInput)
	}
	if len(d.Examples) != 1 || d.Examples[0].Output != "[7,0,8]" {
		t.Errorf("examples = %+v", d.Examples)
	}
}

func TestGetProblemDetail_NotFound(t *testing.T) {
	client := &mockClient{detailResponse: []byte(`{"data":{"question":null}}`)}
	service := NewProfileService(client, "testuser")

	_, err := service.GetProblemDetail(context.Background(), "no-such-problem")
	if err == nil {
		t.Fatal("expected error for null question, got nil")
	}
}

func TestGetProblemDetail_GraphQLError(t *testing.T) {
	client := &mockClient{detailResponse: []byte(`{"errors":[{"message":"Something broke"}]}`)}
	service := NewProfileService(client, "testuser")

	_, err := service.GetProblemDetail(context.Background(), "two-sum")
	if err == nil {
		t.Fatal("expected error for GraphQL errors array, got nil")
	}
	if err.Error() != "Something broke" {
		t.Errorf("error = %q, want %q", err.Error(), "Something broke")
	}
}

func TestGetProblemDetail_NetworkError(t *testing.T) {
	client := &mockClient{detailErr: context.DeadlineExceeded}
	service := NewProfileService(client, "testuser")

	_, err := service.GetProblemDetail(context.Background(), "two-sum")
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
}

func TestParseExamples_FallbackToPreBlocks(t *testing.T) {
	content := `<p>Some description without example class.</p>
<pre>
<strong>Input:</strong> s = "abcabcbb"
<strong>Output:</strong> 3
</pre>`

	examples := parseExamples(content)
	if len(examples) != 1 {
		t.Fatalf("examples count = %d, want 1", len(examples))
	}
	if examples[0].Number != 1 || examples[0].Input != `s = "abcabcbb"` || examples[0].Output != "3" {
		t.Errorf("example = %+v", examples[0])
	}
}

func TestParseSignature_InvalidJSON(t *testing.T) {
	sig := parseSignature("not json")
	if sig.Name != "" || len(sig.Params) != 0 || sig.Return != "" {
		t.Errorf("invalid metaData should yield empty signature, got %+v", sig)
	}
}

func TestParseSignature_ListNotation(t *testing.T) {
	sig := parseSignature(`{"name":"removeInvalidParentheses","params":[{"name":"s","type":"string"},{"name":"seen","type":"map<string, boolean>"}],"return":{"type":"list<string>"}}`)
	if sig.Name != "removeInvalidParentheses" {
		t.Errorf("Name = %q", sig.Name)
	}
	if sig.Return != "string[]" {
		t.Errorf("Return = %q, want %q", sig.Return, "string[]")
	}
	if sig.Params[0].Type != "string" {
		t.Errorf("param[0].Type = %q", sig.Params[0].Type)
	}
	if sig.Params[1].Type != "map<string, boolean>" {
		t.Errorf("param[1].Type = %q", sig.Params[1].Type)
	}
}

func TestNormalizeType(t *testing.T) {
	cases := map[string]string{
		"list<string>":        "string[]",
		"list<integer>":       "integer[]",
		"list<list<integer>>": "integer[][]",
		"integer[]":           "integer[]",
		"string":              "string",
		" ListNode ":          "ListNode",
		"map<string, bool>":   "map<string, bool>",
	}
	for in, want := range cases {
		if got := normalizeType(in); got != want {
			t.Errorf("normalizeType(%q) = %q, want %q", in, got, want)
		}
	}
}
