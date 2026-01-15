/**
 * Convert BBCode and HTML entities to proper HTML/React elements
 */
export function convertBBCode(text: string): string {
  if (!text) return '';

  // Handle HTML entities
  let processed = text
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&');

  // Convert BBCode to HTML
  // URLs
  processed = processed.replace(
    /\[url=([^\]]+)\](.+?)\[\/url\]/gi,
    '<a href="$1" target="_blank" rel="noopener noreferrer">$2</a>'
  );
  processed = processed.replace(
    /\[url\](.+?)\[\/url\]/gi,
    '<a href="$1" target="_blank" rel="noopener noreferrer">$1</a>'
  );

  // Formatting
  processed = processed.replace(/\[b\](.+?)\[\/b\]/gi, '<strong>$1</strong>');
  processed = processed.replace(/\[i\](.+?)\[\/i\]/gi, '<em>$1</em>');
  processed = processed.replace(/\[u\](.+?)\[\/u\]/gi, '<u>$1</u>');
  processed = processed.replace(/\[s\](.+?)\[\/s\]/gi, '<s>$1</s>');

  // Lists
  processed = processed.replace(
    /\[list\](.+?)\[\/list\]/gi,
    (_match, content) => {
      const items = content.split(/\[\*\]/).filter((item: string) => item.trim());
      return `<ul>${items.map((item: string) => `<li>${item.trim()}</li>`).join('')}</ul>`;
    }
  );

  // Quotes
  processed = processed.replace(
    /\[quote(?:=[^\]]*)?\](.+?)\[\/quote\]/gi,
    '<blockquote>$1</blockquote>'
  );

  // Code
  processed = processed.replace(/\[code\](.+?)\[\/code\]/gi, '<code>$1</code>');

  // Images
  processed = processed.replace(
    /\[img\](.+?)\[\/img\]/gi,
    '<img src="$1" style="max-width:100%" loading="lazy" />'
  );

  // Convert newlines to <br>
  processed = processed.replace(/\n/g, '<br>');

  return processed;
}
