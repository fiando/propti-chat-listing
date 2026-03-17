import test from 'node:test';
import assert from 'node:assert/strict';
import { contactSections } from './content.ts';

test('contact page exposes a support email link for platform help', () => {
  const helpSection = contactSections.find((section) => section.title === 'Bantuan penggunaan platform');

  assert.ok(helpSection, 'expected bantuan penggunaan platform section to exist');
  assert.deepEqual(helpSection.supportLink, {
    label: 'Kirim pesan ke support@propti.id',
    href: 'mailto:support@propti.id',
  });
});
