import get from "lodash/get";
import { computed, ref } from "vue";
import { ajv, localizeAjvErrors } from "./ajv";
import myi18n from "./i18n.json";
import type { ValidateFunction } from "ajv";

export const allLocales = ['de', 'en', "fr", "it"];
export type Locale = typeof allLocales[number];
// Reactive locale state
export const locale = ref<Locale>('de')

// Translator factory function
const createTranslator = (loc: Locale) => (key: string, defaultMessage: string | undefined) => {
    //console.log(`${loc}.${key}`, defaultMessage);
    return get(myi18n, `${loc}.${key}`, defaultMessage || "");
}

// Computed translation function based on current locale
export const translation = computed(() => createTranslator(locale.value))

export function compileLocalized<T = unknown>(schema: any): ValidateFunction<T> {
  const validator = ajv.compile<T>(schema);
  const localizedValidator = ((data: any) => {
    const valid = validator(data);
    localizeAjvErrors(validator.errors, locale.value);
    localizedValidator.errors = validator.errors;
    return valid;
  }) as ValidateFunction<T>;

  Object.assign(localizedValidator, validator);

  return localizedValidator;
}


