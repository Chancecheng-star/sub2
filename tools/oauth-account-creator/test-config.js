import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const content = fs.readFileSync(path.join(__dirname, 'config.yaml'), 'utf-8');
console.log('原始内容:');
console.log(content);
console.log('\n---\n');

const lines = content.split('\n');
for (const line of lines) {
  if (line.trim() && !line.trim().startsWith('#')) {
    const [key, value] = line.split(':').map(s => s.trim());
    if (key === 'sub2api_url') {
      console.log('key:', key);
      console.log('value 原始:', value);
      console.log('value 长度:', value.length);
      console.log('value bytes:', [...value].map(c => c.charCodeAt(0)));
      const parsed = value.replace(/^["']|["']$/g, '');
      console.log('value 解析后:', parsed);
    }
  }
}
