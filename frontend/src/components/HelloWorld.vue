<script lang="ts" setup>
import {reactive} from 'vue'
import {Greet} from '@generated/go/main/App'
import {main} from "@generated/go/models";
import {LogError} from "@generated/runtime";

const data = reactive({
  name: "",
  age: 0,
  resultText: "Please enter your name below 👇",
})

function greet() {
  let person = new main.Person();
  person.name = data.name;
  person.age = data.age;
  Greet(person)
      .then(result => {
        data.resultText = result
      })
      .catch(err => {
        LogError(err);

        if (err instanceof Error) {
          data.resultText = err.message;
          return;
        } else if (typeof err === "string") {
          data.resultText = err;
          return;
        }

        data.resultText = "Something went wrong! Please try again";
      })
}

</script>

<template>
  <main>
    <div id="result" class="result">{{ data.resultText }}</div>
    <div id="input" class="input-box">
      <FloatLabel variant="on">
        <InputText id="name-input" type="text" v-model="data.name"/>
        <label for="name-input">Name</label>
      </FloatLabel>
      <FloatLabel variant="on">
        <InputNumber id="age-input" v-model="data.age"/>
        <label for="age-input">Age</label>
      </FloatLabel>
      <Button @click="greet">Greet</Button>
    </div>
  </main>
</template>

<style scoped>
.result {
  height: 20px;
  line-height: 20px;
  margin: 1.5rem auto;
}

.input-box {
  display: flex;
  flex-direction: column;
}

.input-box .input {
  border: none;
  border-radius: 3px;
  outline: none;
  height: 30px;
  line-height: 30px;
  padding: 0 10px;
  background-color: rgba(240, 240, 240, 1);
  -webkit-font-smoothing: antialiased;
}

.input-box .input:hover {
  border: none;
  background-color: rgba(255, 255, 255, 1);
}

.input-box .input:focus {
  border: none;
  background-color: rgba(255, 255, 255, 1);
}
</style>
