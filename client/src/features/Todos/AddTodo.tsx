import { useForm } from "@mantine/form";
import { Button, Group, Modal, Textarea, TextInput } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { KeyedMutator } from "swr";
import { ENDPOINT } from "../../utils/fetcher";
import { Todo } from "./types";

export const AddTodo = ({ mutate }: { mutate: KeyedMutator<Todo[]> }) => {
  const [opened, { open, close }] = useDisclosure(false);

  const form = useForm({
    initialValues: { title: "", body: "" },
  });

  const createTodo = async () => {
    const updated = await fetch(`${ENDPOINT}/api/todos`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(form.values),
    }).then((r) => {
      return r.json();
    });

    await mutate(updated);

    form.reset();
    close();
  };

  return (
    <>
      <Modal opened={opened} onClose={close} title="Create Todo" centered>
        <form onSubmit={form.onSubmit(createTodo)}>
          <TextInput
            required
            mb={12}
            label="Todo"
            placeholder="create todo"
            {...form.getInputProps("title")}
          />
          <Textarea
            required
            mb={12}
            label="Body"
            placeholder="Todo body"
            {...form.getInputProps("body")}
          />
          <Button type="submit">Submit</Button>
        </form>
      </Modal>
      <Group>
        <Button fullWidth mb={12} onClick={open}>
          Open
        </Button>
      </Group>
    </>
  );
};
