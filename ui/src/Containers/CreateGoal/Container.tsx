import React from 'react';

interface GoalForm {
  name: string;
  target: number;
}

export default function CreateGoal() {
  const [error, setError] = React.useState<string>('');
  const [form, setForm] = React.useState<GoalForm>({ name: '', target: 0 });
  return <div></div>;
}
