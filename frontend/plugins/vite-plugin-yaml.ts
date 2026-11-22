import yaml from 'js-yaml';
import type { Plugin } from 'vite';

export default function yamlPlugin(): Plugin {
    return {
        name: 'yaml-to-json',
        transform(code, id) {
            if (!id.endsWith('.yaml') && !id.endsWith('.yml')) return null;
            const parsed = yaml.load(code);
            return { code: `export default ${JSON.stringify(parsed)}`, map: null };
        },
    };
}