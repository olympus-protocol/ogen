export function generateWordList(n = 25) {
  const words = [];

  for (let i = 1; i < n; i++) {
    words.push({ name: `word_${i}`, label: `${i}.` });
  }

  return words;
}
