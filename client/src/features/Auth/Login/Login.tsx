import { useContext, useState } from "react";
import { useForm } from "@mantine/form";
import { TextInput, Button, Modal, Box } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { useToast } from "../../../hooks/useToast";
import { UserContext } from "../../../contexts/UserContext";
import { useNavigate } from "react-router-dom";
import { User } from "../../../types/User";

function Login() {
  const [opened, { open, close }] = useDisclosure(false);
  const navigate = useNavigate();
  const { showToast } = useToast();
  const [formState, setFormState] = useState<"login" | "register">("login");
  const { user, logout, login, register } = useContext(UserContext);

  const { values, errors, isValid, setFieldValue } = useForm({
    initialValues: {
      email: "123@test.com",
      password: "thepassword",
      firstname: "",
      lastname: "",
    },
  });
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (isValid()) {
      let response: User | null = null;
      try {
        let authDto;

        switch (formState) {
          case "login":
            authDto = {
              email: values.email,
              password: values.password,
            };
            response = await login(authDto);
            break;
          case "register":
            authDto = {
              email: values.email,
              password: values.password,
              firstName: values.firstname,
              lastName: values.lastname,
            };
            response = await register(authDto);
            break;
          default:
            break;
        }

        showToast({
          message: "Success",
          title: `Welcome ${response?.firstName} ${response?.lastName}`,
          icon: "check",
          color: "green",
        });
        navigate("/dashboard/todo");
        close();
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } catch (error: any) {
        showToast({
          message: error.message,
          title: "Error",
          icon: "cross",
          color: "red",
        });
      }
    }
  };

  return (
    <>
      <Button onClick={open}>Log in</Button>
      {user && <Button onClick={logout}>Log out</Button>}

      <Modal opened={opened} onClose={close}>
        <form onSubmit={handleSubmit}>
          {formState === "register" && (
            <>
              <TextInput
                label="First Name"
                placeholder="Enter your first name"
                value={values.firstname}
                onChange={(event) =>
                  setFieldValue("firstname", event.currentTarget.value)
                }
                error={errors.firstname}
                required
              />

              <TextInput
                label="Last Name"
                placeholder="Enter your last name"
                value={values.lastname}
                onChange={(event) =>
                  setFieldValue("lastname", event.currentTarget.value)
                }
                error={errors.lastname}
                required
              />
            </>
          )}

          <TextInput
            label="Email"
            placeholder="Enter your email"
            value={values.email}
            onChange={(event) =>
              setFieldValue("email", event.currentTarget.value)
            }
            error={errors.email}
            required
          />

          <TextInput
            label="Password"
            type="password"
            placeholder="Enter your password"
            value={values.password}
            onChange={(event) =>
              setFieldValue("password", event.currentTarget.value)
            }
            error={errors.password}
            required
          />

          <Box
            style={{
              display: "flex",
              justifyContent: "center",
              gap: "12px",
            }}
            mt={12}
          >
            <Button type="submit" disabled={!isValid()}>
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
