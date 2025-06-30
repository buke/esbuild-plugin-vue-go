#!/bin/sh
# Automatically initialize a Vue 3 + TypeScript + Router + Pinia + Sass project

set -e

# Remove any existing vue-example directory
rm -rf vue-example

# Create a new Vue project with TypeScript, Router, Pinia, and default settings
npx -y create-vue vue-example --ts --router --pinia --default
cd vue-example || exit 1

# Install all npm dependencies
npm install

# Create the styles directory if it does not exist
mkdir -p src/styles

# Generate or overwrite src/styles/variables.scss with some base variables
cat > src/styles/variables.scss <<'EOF'
$primary-color: #42b983;
$secondary-color: #35495e;
$font-size-base: 16px;
$border-radius-base: 6px;
$padding-base: 1rem;
EOF

# Generate or overwrite src/styles/index.scss with some global styles
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

# If main.ts does not import index.scss, insert the import after the first import statement
if ! grep -q "import './styles/index.scss'" src/main.ts; then
  # Insert after the first import
  sed -i '' "1a\\
import './styles/index.scss';
" src/main.ts
fi

# If HelloWorld.vue does not contain <style lang="scss" scoped>, append a style block
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