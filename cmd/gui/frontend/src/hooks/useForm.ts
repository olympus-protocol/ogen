import { useState } from 'react';

export default function useForm() {
  const [form, setForm] = useState({});
  const handleTextInput = (e: any) =>
    setForm({ ...form, [e.target.name]: e.target.value });

  return { form, handleTextInput, setForm };
}
