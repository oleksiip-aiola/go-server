import { useState } from "react";
import { useForm } from "@mantine/form";
import { TextInput, Button, Modal, Box } from "@mantine/core";
import { ENDPOINT } from "../../../App";
import { useDisclosure } from "@mantine/hooks";
import { useToast } from "../../../hooks/useToast";

function Login() {
  const [opened, { open, close }] = useDisclosure(false);
  const [error, setError] = useState("");
  const { showToast } = useToast();
  const [formState, setFormState] = useState<"login" | "register">("login");

  const form = useForm({
    initialValues: {
      email: "123@test.com",
      password: "thepassword",
      firstname: "",
      lastname: "",
    },
  });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (form.isValid()) {
      const endpoint = formState === "login" ? "/api/login" : "/api/register";
      await fetch(`${ENDPOINT}${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: form.values.email,
          password: form.values.password,
          firstname: form.values.firstname,
          lastname: form.values.lastname,
        }),
      }).then(async (r) => {
        if (!r.ok) {
          const x = await r.json();
          setError(x.detail);
          showToast({
            message: x.detail,
            title: "Error",
            icon: "cross",
            color: "red",
          });
          throw new Error("Request failed with status " + r.status);
        }
        showToast({
          message: "Success",
          title: "Success",
          icon: "check",
          color: "green",
        });

        r.json();
        close();
      });
    }
  };

  return (
    <>
      <Button onClick={open}>Log in</Button>

      <Modal opened={opened} onClose={close}>
        <form onSubmit={handleSubmit}>
          {formState === "register" && (
            <>
              <TextInput
                label="First Name"
                placeholder="Enter your first name"
                value={form.values.firstname}
                onChange={(event) =>
                  form.setFieldValue("firstname", event.currentTarget.value)
                }
                error={form.errors.firstname}
                required
              />

              <TextInput
                label="Last Name"
                placeholder="Enter your last name"
                value={form.values.lastname}
                onChange={(event) =>
                  form.setFieldValue("lastname", event.currentTarget.value)
                }
                error={form.errors.lastname}
                required
              />
            </>
          )}

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

          {error && (
            <Box mt={4} style={{ color: "red" }}>
              {error}
            </Box>
          )}

          <Box
            style={{
              display: "flex",
              justifyContent: "center",
              gap: "12px",
            }}
            mt={12}
          >
            <Button type="submit" disabled={!form.isValid()}>
              {formState === "login" ? "Login" : "Register"}
            </Button>

            <Button
              ml={"auto"}
              size="xs"
              onClick={() =>
                setFormState(formState === "login" ? "register" : "login")
              }
            >
              {formState === "login" ? "No account? Sign up" : "Sign in"}
            </Button>
          </Box>
        </form>
      </Modal>
    </>
  );
}

export default Login;
