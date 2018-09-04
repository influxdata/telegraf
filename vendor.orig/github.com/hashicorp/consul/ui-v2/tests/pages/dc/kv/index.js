export default function(visitable, deletable, creatable, clickable, attribute, collection) {
  return creatable({
    visit: visitable('/:dc/kv'),
    kvs: collection(
      '[data-test-tabular-row]',
      deletable({
        name: attribute('data-test-kv', '[data-test-kv]'),
        kv: clickable('a'),
        actions: clickable('label'),
      })
    ),
  });
}
