#include <stdio.h>
#include <stdlib.h>
typedef struct {
	int data; // 数据节点
	struct TreeNode *left; // 指向左子树
	struct TreeNode *right; // 指向右子树
} TreeNode , *PTreeNode;

// 记录平衡二叉树
bool BalanceTrue = false;
// 最小不平衡子树地址
TreeNode *rjt = NULL;

// 检查二叉树是否平衡，若不平衡 BalanceTrue 为 true
int checkTreeBalance(TreeNode *root) {
	if (NULL == root) { return 0; }
	int x = checkTreeBalance(root->left);
	int y = checkTreeBalance(root->right);

	// 若检测到最小不平衡二叉树后，不进行后面的检查
	if (BalanceTrue) return 0;

	int xx = abs(x-y);

	if (xx > 1) {
			// 左子树 和 右子树 相差大于1 ， 二叉树不平衡
			BalanceTrue = true;
			rjt = root;
	}
	 
	return (x>y?x+1:y+1);
}
