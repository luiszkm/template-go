export function groupByFeature(permissions: string[]) {
  const groups = new Map<string, string[]>();
  for (const permission of [...permissions].sort()) {
    const feature = permission.split(":")[0] ?? permission;
    groups.set(feature, [...(groups.get(feature) ?? []), permission]);
  }
  return [...groups.entries()];
}

export function PermissionPicker({
  catalogue,
  checked,
  disabled,
  error,
  onChange,
}: {
  catalogue: string[];
  checked: string[];
  disabled?: boolean;
  error?: string;
  onChange: (checked: string[]) => void;
}) {
  function toggle(permission: string) {
    onChange(
      checked.includes(permission)
        ? checked.filter((p) => p !== permission)
        : [...checked, permission].sort(),
    );
  }
  return (
    <div className="space-y-3" aria-describedby={error ? "permissions-error" : undefined}>
      <p className="text-sm font-medium">Permissões</p>
      {groupByFeature(catalogue).map(([feature, permissions]) => (
        <div key={feature} className="space-y-1">
          <h3 className="text-sm font-semibold">{feature}</h3>
          {permissions.map((permission) => (
            <label key={permission} className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={checked.includes(permission)}
                disabled={disabled}
                onChange={() => toggle(permission)}
              />
              {permission}
            </label>
          ))}
        </div>
      ))}
      {error && (
        <p id="permissions-error" className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
