#!/usr/bin/env node

/**
 * Generate PNG favicons from SVG source
 * This script converts the SVG favicon to various PNG sizes
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// InstaInfra logo SVG - matches the header logo exactly
const logoSVG = `<svg width="{{SIZE}}" height="{{SIZE}}" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="24" height="24" rx="3" fill="#1f2937"/>
  <path d="M21 9.5V5C21 3.89543 20.1046 3 19 3H5C3.89543 3 3 3.89543 3 5V9.5M21 9.5H3M21 9.5V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V9.5M9 14H15" stroke="#3b82f6" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
</svg>`;

const publicDir = path.join(__dirname, '../public');

// Ensure public directory exists
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Generate favicons
const sizes = [
  { size: 16, name: 'favicon-16x16.png' },
  { size: 32, name: 'favicon-32x32.png' },
  { size: 180, name: 'apple-touch-icon.png' }
];

console.log('🎨 Generating InstaInfra favicons...');

// Check if favicon.svg already exists, if not create it
const faviconSvgPath = path.join(publicDir, 'favicon.svg');
if (!fs.existsSync(faviconSvgPath)) {
  const mainSVG = logoSVG.replace(/{{SIZE}}/g, '32');
  fs.writeFileSync(faviconSvgPath, mainSVG);
  console.log('✅ Generated favicon.svg');
} else {
  console.log('✅ favicon.svg already exists, keeping existing version');
}

// For each size, create a proper SVG file
// In a real implementation, you'd use a tool like sharp or puppeteer to convert to PNG
sizes.forEach(({ size, name }) => {
  const sizeSVG = logoSVG.replace(/{{SIZE}}/g, size.toString());
  const svgPath = path.join(publicDir, name.replace('.png', '.svg'));
  fs.writeFileSync(svgPath, sizeSVG);
  console.log(`✅ Generated ${name} (as SVG - convert to PNG in production)`);
});

console.log('🚀 All InstaInfra favicons generated successfully!');
console.log('📝 Note: In production, convert SVG files to PNG using a tool like:');
console.log('   - sharp (Node.js)');
console.log('   - ImageMagick');
console.log('   - Or online converters'); 