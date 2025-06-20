import Ajv from 'ajv';
import localize from 'ajv-i18n';

export const ajv = new Ajv({ allErrors: true });

export function localizeAjvErrors(errors: any[] | null | undefined, locale: string) {
  if (!errors) return;

  const localizer = (localize as any)[locale];
  if (typeof localizer === 'function') {
    localizer(errors); // modifies in place
  }
}
