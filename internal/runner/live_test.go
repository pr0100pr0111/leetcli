package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"leetcli/internal/api"
	"leetcli/internal/domain"
	"leetcli/internal/service"
)

var leetSlugForLang = map[string]string{
	"go":         "golang",
	"python":     "python3",
	"javascript": "javascript",
	"cpp":        "cpp",
}

const liveSrcRoot = "/tmp/leetcli-live-src"

const addTwoNumbersPython = `class ListNode:
    def __init__(self, val=0, next=None):
        self.val = val
        self.next = next

class Solution:
    def addTwoNumbers(self, l1, l2):
        dummy = ListNode()
        cur = dummy
        carry = 0
        while l1 or l2 or carry:
            a = l1.val if l1 else 0
            b = l2.val if l2 else 0
            s = a + b + carry
            carry = s // 10
            cur.next = ListNode(s % 10)
            cur = cur.next
            l1 = l1.next if l1 else None
            l2 = l2.next if l2 else None
        return dummy.next
`

const addTwoNumbersGo = `type ListNode struct {
	Val  int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	cur := dummy
	carry := 0
	for l1 != nil || l2 != nil || carry > 0 {
		a, b := 0, 0
		if l1 != nil {
			a = l1.Val
			l1 = l1.Next
		}
		if l2 != nil {
			b = l2.Val
			l2 = l2.Next
		}
		s := a + b + carry
		carry = s / 10
		cur.Next = &ListNode{Val: s % 10}
		cur = cur.Next
	}
	return dummy.Next
}`

const addTwoNumbersJS = `class ListNode {
    constructor(val, next) {
        this.val = (val === undefined ? 0 : val);
        this.next = (next === undefined ? null : next);
    }
}
class Solution {
    addTwoNumbers(l1, l2) {
        const dummy = new ListNode();
        let cur = dummy;
        let carry = 0;
        while (l1 || l2 || carry) {
            const a = l1 ? l1.val : 0;
            const b = l2 ? l2.val : 0;
            const s = a + b + carry;
            carry = Math.floor(s / 10);
            cur.next = new ListNode(s % 10);
            cur = cur.next;
            if (l1) l1 = l1.next;
            if (l2) l2 = l2.next;
        }
        return dummy.next;
    }
}`

const addTwoNumbersCpp = `struct ListNode {
    int val;
    ListNode *next;
    ListNode() : val(0), next(nullptr) {}
    ListNode(int x) : val(x), next(nullptr) {}
    ListNode(int x, ListNode *next) : val(x), next(next) {}
};

class Solution {
public:
    ListNode* addTwoNumbers(ListNode* l1, ListNode* l2) {
        ListNode dummy;
        ListNode* cur = &dummy;
        int carry = 0;
        while (l1 || l2 || carry) {
            int a = l1 ? l1->val : 0;
            int b = l2 ? l2->val : 0;
            int s = a + b + carry;
            carry = s / 10;
            cur->next = new ListNode(s % 10);
            cur = cur->next;
            if (l1) l1 = l1->next;
            if (l2) l2 = l2->next;
        }
        return dummy.next;
    }
};`

func liveDetail(t *testing.T, slug string) domain.ProblemDetail {
	t.Helper()
	svc := service.NewProfileService(api.NewClient(), "pr0100pr0")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	detail, err := svc.GetProblemDetail(ctx, slug)
	if err != nil {
		t.Fatalf("fetch %q: %v", slug, err)
	}
	return detail
}

func runLive(t *testing.T, detail domain.ProblemDetail, lang, solution string) {
	t.Helper()
	task, err := NewTask(detail, leetSlugForLang[lang])
	if err != nil {
		t.Errorf("NewTask(%s): %v", lang, err)
		return
	}
	task.Solution = solution
	r, _ := For(lang)
	dir := filepath.Join(liveSrcRoot, detail.Slug, lang)
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("clean %s: %v", dir, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results, err := r.Run(ctx, task, dir)
	if err != nil {
		t.Errorf("run %s: %v", lang, err)
		return
	}
	res := results[0]
	t.Logf("[%s/%s] status=%s got=%q want=%q msg=%q", detail.Slug, lang, res.Status, res.Got, res.Want, res.Message)
	if res.Status != StatusPass {
		t.Errorf("[%s/%s] expected PASS, got %s", detail.Slug, lang, res.Status)
	}
}

func runLiveAll(t *testing.T, detail domain.ProblemDetail, solutions map[string]string) {
	t.Helper()
	ran := 0
	for _, lang := range Languages() {
		r, _ := For(lang)
		if !r.Available() {
			t.Logf("skip %s: toolchain not available", lang)
			continue
		}
		runLive(t, detail, lang, solutions[lang])
		ran++
	}
	if ran == 0 {
		t.Fatal("no toolchains available")
	}
}

func TestLiveTwoSum(t *testing.T) {
	if os.Getenv("LEETCLI_RUNNER_LIVE") == "" {
		t.Skip("set LEETCLI_RUNNER_LIVE=1")
	}
	detail := liveDetail(t, "two-sum")
	if detail.Signature.Name != "twoSum" {
		t.Fatalf("signature = %+v", detail.Signature)
	}
	runLiveAll(t, detail, map[string]string{
		"go":         twoSumGo,
		"python":     twoSumPython,
		"javascript": twoSumJS,
		"cpp":        twoSumCpp,
	})
}

func TestLiveAddTwoNumbers(t *testing.T) {
	if os.Getenv("LEETCLI_RUNNER_LIVE") == "" {
		t.Skip("set LEETCLI_RUNNER_LIVE=1")
	}
	detail := liveDetail(t, "add-two-numbers")
	if detail.Signature.Name != "addTwoNumbers" {
		t.Fatalf("signature = %+v", detail.Signature)
	}
	runLiveAll(t, detail, map[string]string{
		"go":         addTwoNumbersGo,
		"python":     addTwoNumbersPython,
		"javascript": addTwoNumbersJS,
		"cpp":        addTwoNumbersCpp,
	})
}
