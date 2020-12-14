import Vue from 'vue';

import hljs from 'highlight.js/lib/core';
import 'highlight.js/styles/github.css';
import { getDefinition } from '@/utils/language.js';

hljs.registerLanguage('goxy', getDefinition);

Vue.use(hljs.vuePlugin);
