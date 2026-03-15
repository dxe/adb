export type NavAccessRules = {
  roleRequired?: string[];
  visibleForNonSFBay?: boolean;
};

export type NavItem = NavAccessRules & {
  label: string;
  href: string;
  page: string;
  separatorBelow?: boolean;
};

export type NavDropdownItem = NavAccessRules & {
  label: string;
  items: NavItem[];
};

export type NavbarData = {
  items: NavDropdownItem[];
};

export function userHasNavRole(userRoles: string[], role: string): boolean {
  if (role === "admin") {
    return userRoles.includes("admin");
  }
  if (role === "organizer") {
    return userRoles.includes("admin") || userRoles.includes("organizer");
  }
  if (role === "attendance") {
    return (
      userRoles.includes("admin") ||
      userRoles.includes("organizer") ||
      userRoles.includes("attendance")
    );
  }
  return userRoles.includes(role);
}

export function evaluateNavAccess(
  userRoles: string[],
  chapterId: number,
  item: NavAccessRules,
  sfBayChapterId: number,
  parentItem?: NavAccessRules,
): boolean {
  const visibleForNonSFBay =
    item.visibleForNonSFBay !== undefined
      ? item.visibleForNonSFBay
      : parentItem && parentItem.visibleForNonSFBay !== undefined
        ? parentItem.visibleForNonSFBay
        : false;

  if (chapterId !== sfBayChapterId && !visibleForNonSFBay) {
    return false;
  }

  if (!item.roleRequired) {
    return true;
  }

  if (!userRoles.length) {
    return false;
  }

  return item.roleRequired.some((role) => userHasNavRole(userRoles, role));
}
