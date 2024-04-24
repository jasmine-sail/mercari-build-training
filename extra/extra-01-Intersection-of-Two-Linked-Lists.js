var getIntersectionNode = function(headA, headB) {
    let lenA = getLength(headA);
    let lenB = getLength(headB);
    
    while (lenA > lenB) {
        headA = headA.next;
        lenA--;
    }
    while (lenB > lenA) {
        headB = headB.next;
        lenB--;
    }
    while (headA !== null && headB !== null) {
        if (headA === headB) {
            return headA;
        }
        headA = headA.next;
        headB = headB.next;
    }
    return null;
};


function getLength(head) {
    let length = 0;
    while (head !== null) {
        length++;
        head = head.next;
    }
    return length;
};