import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import { setupPermissionDirective } from './directives/permission';
import './styles/tokens.css';
import './styles/global.css';
import './styles/page.css';

/*
 * 全局组件注册 —— vue-login-skill 必备
 *
 * 背景：vue-login-skill 的 LoginForm.vue 内部用 <base-card> <base-form> <base-input> <base-button>
 *   <base-checkbox> 等标签（"零 HTML5 标签铁律"），但**不显式 import**——假设消费者把它们注册为
 *   全局组件（因为 LoginForm.vue 在 templates/ 下无法用相对路径 import 软链接包）。
 *
 * 不注册会怎样：LoginForm.vue 渲染出的 DOM 含字面 <base-input> <base-form> 等 self-closing 标签
 *   —— 浏览器把它们当未知 HTML 元素忽略，input/button 全不见，只有标题可见。
 *
 * 一次性注册全部 base-* 原子组件 —— 既满足 vue-login-skill 的需求，又能让未来任何包装组件
 *   （FlipLoginContainer / SplitLayout 等）直接用 base-* 不必重复 import。
 */
import { BaseCard } from 'vue-card-skill'
import { BaseForm, BaseFormItem } from 'vue-form-skill'
import { BaseInput } from 'vue-input-skill'
import { BaseButton } from 'vue-button-skill'
import { BaseCheckbox } from 'vue-checkbox-skill'

const app = createApp(App)

app.component('BaseCard', BaseCard)
app.component('BaseForm', BaseForm)
app.component('BaseFormItem', BaseFormItem)
app.component('BaseInput', BaseInput)
app.component('BaseButton', BaseButton)
app.component('BaseCheckbox', BaseCheckbox)

app.use(createPinia());
app.use(router);
setupPermissionDirective(app);

app.mount('#app');
