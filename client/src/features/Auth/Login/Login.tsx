import { useForm } from "@mantine/form";
import { TextInput, Button } from "@mantine/core";
import { ENDPOINT } from "../../../App";

function Login() {
  const form = useForm({
    initialValues: {
      email: "123@test.com",
      password: "thepassword",
    },
  });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (form.isValid()) {
      await fetch(`${ENDPOINT}/api/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: form.values.email,
          password: form.values.password,
        }),
      }).then((r) => {
        return r.json();
      });
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <TextInput
        label="Email"
        placeholder="Enter your email"
        value={form.values.email}
        onChange={(event) =>
          form.setFieldValue("email", event.currentTarget.value)
        }
        error={form.errors.email}
        required
      />

      <TextInput
        label="Password"
        type="password"
        placeholder="Enter your password"
        value={form.values.password}
        onChange={(event) =>
          form.setFieldValue("password", event.currentTarget.value)
        }
        error={form.errors.password}
        required
      />

      <Button type="submit" disabled={!form.isValid()}>
        Login
      </Button>
    </form>
  );
}

export default Login;
