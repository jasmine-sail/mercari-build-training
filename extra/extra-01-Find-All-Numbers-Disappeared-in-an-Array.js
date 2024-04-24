var findDisappearedNumbers = function(nums) {
    const n = nums.length;
    const numsSet = new Set(nums);
    const result = [];

    for (let i = 1; i <= n; i++) {
        if (!numsSet.has(i)) {
            result.push(i);
        }
    }

    return result;
};