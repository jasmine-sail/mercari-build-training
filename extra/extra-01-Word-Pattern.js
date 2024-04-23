var wordPattern = function(pattern, s) {
    const words = s.split(" ");
    
    if (pattern.length !== words.length) {
        return false;
    }
    
    const map1 = new Map(); 
    const map2 = new Map(); 
    
    for (let i = 0; i < pattern.length; i++) {
        const char = pattern[i];
        const word = words[i];
        
        if (!map1.has(char)) {
            map1.set(char, word);
        } else {
            if (map1.get(char) !== word) {
                return false;
            }
        }
        
        if (!map2.has(word)) {
            map2.set(word, char);
        } else {
            if (map2.get(word) !== char) {
                return false;
            }
        }
    }
    
    return true;
    };