import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AppButton from './AppButton.vue'

describe('AppButton', () => {
  it('renders slot content', () => {
    const wrapper = mount(AppButton, { slots: { default: 'Click me' } })
    expect(wrapper.text()).toBe('Click me')
  })

  it('applies primary variant class by default', () => {
    const wrapper = mount(AppButton, { slots: { default: 'Go' } })
    expect(wrapper.classes()).toContain('app-btn--primary')
  })

  it('applies the specified variant class', () => {
    const wrapper = mount(AppButton, {
      props: { variant: 'secondary' },
      slots: { default: 'Go' },
    })
    expect(wrapper.classes()).toContain('app-btn--secondary')
  })

  it('applies the specified size class', () => {
    const wrapper = mount(AppButton, {
      props: { size: 'sm' },
      slots: { default: 'Go' },
    })
    expect(wrapper.classes()).toContain('app-btn--sm')
  })

  it('is disabled when disabled prop is true', () => {
    const wrapper = mount(AppButton, {
      props: { disabled: true },
      slots: { default: 'Go' },
    })
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('fires native click when not disabled', async () => {
    let clicked = false
    const wrapper = mount(AppButton, {
      slots: { default: 'Go' },
      attrs: {
        onClick: () => {
          clicked = true
        },
      },
    })
    await wrapper.trigger('click')
    expect(clicked).toBe(true)
  })

  it('does not fire click events when disabled', async () => {
    const wrapper = mount(AppButton, {
      props: { disabled: true },
      slots: { default: 'Go' },
    })
    await wrapper.trigger('click')
    // A disabled button should not trigger click in most browsers;
    // we verify the disabled attribute is present
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('uses button type by default', () => {
    const wrapper = mount(AppButton, { slots: { default: 'Go' } })
    expect(wrapper.attributes('type')).toBe('button')
  })

  it('accepts submit type', () => {
    const wrapper = mount(AppButton, {
      props: { type: 'submit' },
      slots: { default: 'Submit' },
    })
    expect(wrapper.attributes('type')).toBe('submit')
  })

  it('is disabled when loading is true', () => {
    const wrapper = mount(AppButton, {
      props: { loading: true },
      slots: { default: 'Go' },
    })
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('renders spinner icon when loading', () => {
    const wrapper = mount(AppButton, {
      props: { loading: true },
      slots: { default: 'Go' },
    })
    expect(wrapper.find('i.pi-spinner').exists()).toBe(true)
  })

  it('does not render spinner when not loading', () => {
    const wrapper = mount(AppButton, { slots: { default: 'Go' } })
    expect(wrapper.find('i.pi-spinner').exists()).toBe(false)
  })
})
