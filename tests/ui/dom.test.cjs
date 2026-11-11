// Run with JSDOM_MODULE pointing to an external jsdom install (no app dependency).
// Render fixtures first: go run tests/ui/render.go <fixture-directory>.
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { JSDOM } = require(process.env.JSDOM_MODULE || 'jsdom');
const fixtureDir = process.env.UI_FIXTURES;
const script = name => fs.readFileSync(`static/js/${name}.js`, 'utf8');
function dom(html = '') { return new JSDOM(html, { runScripts: 'outside-only', url: 'https://shop.test' }); }

test('all 11 rendered pages have valid inline JS, unique IDs, and valid label targets', () => {
 const files = fs.readdirSync(fixtureDir).filter(f => f.endsWith('.html'));
 assert.equal(files.length, 11);
 for (const file of files) {
  const d = dom(fs.readFileSync(path.join(fixtureDir, file), 'utf8'));
  const doc = d.window.document;
  const ids = [...doc.querySelectorAll('[id]')].map(e => e.id);
  assert.equal(new Set(ids).size, ids.length, file + ': duplicate IDs');
  for (const label of doc.querySelectorAll('label[for]')) assert.ok(label.control, file + ': unbound label');
  for (const s of doc.querySelectorAll('script:not([src])')) new Function(s.textContent);
  if (file === 'products.html') assert.ok(doc.getElementById('search-input').closest('.toolbar'));
  assert.ok(!doc.querySelector('meta[name="viewport"]').content.includes('user-scalable=no'));
  d.window.close();
 }
});

test('validation isolates sibling errors and preserves helper descriptions', () => {
 const d = dom('<form><div><span id="hint">Ayuda</span><input id="a" required aria-describedby="hint"><input id="b" required></div></form>');
 d.window.eval(script('validator'));
 const { document: doc, FormValidator: v } = d.window;
 const a=doc.getElementById('a'), b=doc.getElementById('b');
 assert.equal(v.validateForm(doc.querySelector('form')), false);
 assert.equal(doc.activeElement, a);
 assert.equal(doc.querySelectorAll('.validation-error').length, 2);
 assert.equal(a.getAttribute('aria-invalid'), 'true');
 a.value='Correcto'; v.validateField(a);
 assert.equal(a.getAttribute('aria-describedby'), 'hint');
 assert.equal(a.hasAttribute('aria-invalid'), false);
 assert.equal(doc.querySelectorAll('.validation-error').length, 1);
 assert.equal(b.getAttribute('aria-invalid'), 'true');
 d.window.close();
});

test('toast treats untrusted text literally and exposes status and dismiss name', () => {
 const d = dom(); d.window.eval(script('toast')+'\nwindow.TestToast = Toast;');
 const t = d.window.TestToast.show('error', '<img src=x>', '<script>bad()</script>', 0);
 assert.equal(t.querySelector('img,script'), null);
 assert.equal(t.querySelector('.toast-title').textContent, '<img src=x>');
 assert.equal(t.getAttribute('role'), 'alert');
 assert.ok(t.querySelector('button').getAttribute('aria-label'));
 t.querySelector('button').click(); assert.equal(t.isConnected, false);
 d.window.close();
});

test('confirmation cancels by default, escapes content, resolves once and restores focus', async () => {
 const d = dom('<button id="opener">Abrir</button>'); const w=d.window;
 // jsdom has no modal top layer: these stubs do NOT verify native focus trapping.
 w.HTMLDialogElement.prototype.showModal=function(){this.open=true;};
 w.HTMLDialogElement.prototype.close=function(){this.open=false;};
 const source=fs.readFileSync('views/layout/base.templ','utf8');
 w.eval(source.slice(source.indexOf('window.confirmDialog ='),source.lastIndexOf('</script>')));
 const opener=w.document.getElementById('opener'); opener.focus();
 const result=w.confirmDialog({title:'<img src=x>', message:'<b>Literal</b>'});
 const dialog=w.document.querySelector('dialog');
 assert.ok(dialog.open); assert.equal(dialog.querySelector('img,b'),null);
 const cancel=dialog.querySelector('[id$="-cancel"]'); assert.equal(w.document.activeElement,cancel);
 cancel.click(); assert.equal(await result,false); assert.equal(w.document.activeElement,opener);
 const escape=w.confirmDialog(); w.document.querySelector('dialog').dispatchEvent(new w.Event('cancel',{cancelable:true}));
 assert.equal(await escape,false);
 const accept=w.confirmDialog(); w.document.querySelector('[id$="-ok"]').click(); assert.equal(await accept,true);
 d.window.close();
});

test('user disclosure exposes state and closes with Escape', () => {
 const d=dom(fs.readFileSync(path.join(fixtureDir,'customers.html'),'utf8'));
 const doc=d.window.document;
 const code=[...doc.querySelectorAll('script')].find(s=>s.textContent.includes("var trigger  ="));
 d.window.eval(code.textContent);
 const button=doc.getElementById('user-menu-trigger'); assert.equal(button.tagName,'BUTTON');
 button.click(); assert.equal(button.getAttribute('aria-expanded'),'true');
 button.dispatchEvent(new d.window.KeyboardEvent('keydown',{key:'Escape',bubbles:true}));
 assert.equal(button.getAttribute('aria-expanded'),'false'); assert.equal(doc.activeElement,button);
 d.window.close();
});

test('secondary text tokens reach 4.5:1 on defined static surfaces', () => {
 const luminance=hex=>{const rgb=hex.match(/\w\w/g).map(v=>parseInt(v,16)/255).map(v=>v<=.04045?v/12.92:((v+.055)/1.055)**2.4); return rgb[0]*.2126+rgb[1]*.7152+rgb[2]*.0722;};
 const css=fs.readFileSync('static/css/app-main.css','utf8');
 const themes=[css.match(/:root\s*\{([^}]+)\}/)[1],css.match(/\[data-theme='light'\]\s*\{([^}]+)\}/)[1]];
 for (const theme of themes) {
  const token=name=>theme.match(new RegExp('--'+name+':\\s*#([0-9a-f]{6})'))[1];
  const fg=token('text-mut'), surfaces=['bg','bg-elev','bg-surf','bg-hover'].map(token);
  for (const bg of surfaces) {const values=[luminance(fg),luminance(bg)].sort((a,b)=>b-a); const ratio=(values[0]+.05)/(values[1]+.05); assert.ok(ratio>=4.5, `${fg}/${bg}: ${ratio}`);}
 }
});
