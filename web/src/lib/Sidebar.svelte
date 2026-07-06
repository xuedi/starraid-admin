<script>
  import Icon from './Icon.svelte'
  import { route, section } from './router.svelte.js'

  const items = [
    { name: 'dashboard', href: '#/', label: 'Dashboard', icon: 'dashboard' },
    { name: 'users', href: '#/users', label: 'Users', icon: 'users' },
    { name: 'objects', href: '#/objects', label: 'Objects', icon: 'objects' },
    { name: 'map', href: '#/map', label: 'Map', icon: 'map' },
  ]

  let active = $derived(section(route.name))
</script>

<aside class="bg-base-200 text-base-content min-h-full w-64 flex flex-col border-r border-base-300">
  <div class="px-5 py-4 border-b border-base-300">
    <div class="text-lg font-semibold tracking-wide">StarRaid</div>
    <div class="text-xs opacity-60">Game-Master Console</div>
  </div>

  <ul class="menu menu-md gap-1 px-3 py-4 grow">
    {#each items as item}
      <li>
        <a
          href={item.href}
          class:menu-active={active === item.name}
          aria-current={active === item.name ? 'page' : undefined}
        >
          <Icon name={item.icon} size={18} />
          {item.label}
        </a>
      </li>
    {/each}
  </ul>

  <div class="px-5 py-3 text-[0.7rem] leading-snug opacity-50 border-t border-base-300">
    Read-only monitor · v0
  </div>
</aside>
