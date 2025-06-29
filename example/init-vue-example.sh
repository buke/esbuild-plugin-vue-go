#!/bin/sh
# 自动初始化 Vue3+TypeScript+Router+Pinia+Sass 项目

set -e

rm -rf vue-example
npx -y create-vue vue-example --ts --router --pinia --default
cd vue-example || exit 1
npm install
npm install -D sass


mkdir -p src/styles

# 生成/覆盖 src/styles/variables.scss
cat > src/styles/variables.scss <<'EOF'
$primary-color: #42b983;
$secondary-color: #35495e;
$font-size-base: 16px;
$border-radius-base: 6px;
$padding-base: 1rem;
EOF

# 生成/覆盖 src/styles/index.scss
cat > src/styles/index.scss <<'EOF'
@import './variables.scss';

body {
  color: $primary-color;
  font-size: $font-size-base;
  background: #f8f8f8;
  margin: 0;
  padding: 0;
}

a {
  color: $secondary-color;
  text-decoration: none;
  transition: color 0.2s;
}

a:hover {
  color: $primary-color;
}
EOF

# 如果 main.ts 没有引入 index.scss，则插入
if ! grep -q "import './styles/index.scss'" src/main.ts; then
  # 在第一个 import 之后插入
  sed -i '' "1a\\
import './styles/index.scss';
" src/main.ts
fi

# 如果 HelloWorld.vue 没有 <style lang="scss" scoped>，则插入
if ! grep -q "<style lang=\"scss\" scoped>" src/components/HelloWorld.vue; then
  cat <<'EOF' >> src/components/HelloWorld.vue

<style lang="scss" scoped>
@import '../styles/variables.scss';

.greetings {
  color: $primary-color;
  font-size: $font-size-base;
  border: 2px dashed $primary-color;
  padding: $padding-base;
  border-radius: $border-radius-base;

  .primary-btn {
    background: $primary-color;
    color: #fff;
    border: none;
    border-radius: $border-radius-base;
    padding: 0.5em 1.5em;
    cursor: pointer;
    font-size: 1em;
    margin-top: 1em;
    transition: background 0.2s;

    &:hover {
      background: $secondary-color;
    }
  }
}
</style>
EOF
fi