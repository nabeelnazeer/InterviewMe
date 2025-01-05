'use client';
import { useEffect } from 'react';

export default function ClientEffects() {
  useEffect(() => {
    document.body.removeAttribute('data-new-gr-c-s-check-loaded');
    document.body.removeAttribute('data-gr-ext-installed');
  }, []);

  return null;
}
