<template>
  <div>
    <v-toolbar flat>
      <v-toolbar-title>Domain - {{ currentCluster }}</v-toolbar-title>
      <v-divider
        class="mx-2"
        inset
        vertical
      />
      <v-spacer/>
      <v-text-field
        v-model="search"
        append-icon="mdi-magnify"
        label="Search"
        single-line
        hide-details
      />
      <v-menu offset-y>
        <template v-slot:activator="{ on }">
          <v-btn v-on="on" color="primary" text>Domain</v-btn>
        </template>
        <v-list>
          <v-subheader>SELECT DOMAIN</v-subheader>
          <v-list-item
            v-for="item in clusters"
            :key="item"
            @click="getConfig(item)"
          >
            <!-- <v-list-item-icon>
              <v-icon>mdi-buffer</v-icon>
            </v-list-item-icon> -->
            <v-list-item-content>
              <v-list-item-title v-text="item" />
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
      <v-dialog v-model="dialog.item" persistent max-width="500px">
        <template v-slot:activator="{ on }">
          <v-btn v-on="on" color="primary" text>New Item</v-btn>
        </template>
        <v-form ref="form" v-model="valid" lazy-validation>
          <v-card>
            <v-card-title>
              <span class="headline">{{ formTitle }} - {{ currentCluster }}</span>
            </v-card-title>
            <v-card-text>
              <v-container grid-list-md>
                <v-layout wrap>
                  <v-flex xs12 sm6>
                    <v-combobox v-model="editedItem.file" :items="files" :readonly="!editable" :rules="emptyRules" label="File name" required/>
                  </v-flex>
                  <v-flex xs12 sm6>
                    <v-text-field v-model="editedItem.section" :readonly="!editable" :rules="emptyRules" label="Section" required/>
                  </v-flex>
                  <v-flex xs12 sm6>
                    <v-text-field v-model="editedItem.key" :readonly="!editable" :rules="emptyRules" label="Key" required/>
                  </v-flex>
                  <v-flex xs12 sm6>
                    <v-select v-model="editedItem.type" :items="keyTypes" :readonly="!editable" :rules="emptyRules" label="Type" required/>
                  </v-flex>
                  <v-flex xs12>
                    <v-textarea v-model="editedItem.value" :readonly="!editable" label="Value"/>
                  </v-flex>
                  <v-flex v-if="editable" xs12>
                    <v-text-field v-model="editedItem.note" :rules="emptyRules" label="Comment" required/>
                  </v-flex>
                </v-layout>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer/>
              <v-btn color="red darken-1" text @click.native="close">Cancel</v-btn>
              <v-btn v-if="editable" color="blue darken-1" text @click.native="save">Save</v-btn>
            </v-card-actions>
          </v-card>
        </v-form>
      </v-dialog>



      <v-dialog v-model="dialog.domain" persistent max-width="500px">
        <template v-slot:activator="{ on }">
          <v-btn v-on="on" color="primary" text>New Domain</v-btn>
        </template>
        <v-card>
          <v-form ref="domainForm" v-model="valid" lazy-validation>
            <v-card-title>
              <span class="headline">New Domain</span>
            </v-card-title>
            <v-card-text>
              <v-container grid-list-md>
                <v-layout wrap>
                  <v-flex xs12 sm6>
                    <v-text-field
                            v-model="newDomain.name"
                            :rules="emptyRules"
                            label="Domain Name"
                            required
                    />
                  </v-flex>
                  <v-flex xs12>
                    <v-file-input
                            v-model="newDomain.files"
                            :rules="emptyRules"
                            accept=".conf"
                            show-size
                            counter
                            multiple
                            label="Config Files (*.conf)"
                    />
                  </v-flex>
                </v-layout>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer />
              <v-btn color="red darken-1" text @click="close()">Cancel</v-btn>
              <v-btn color="blue darken-1" text @click="saveDomain">Save</v-btn>
            </v-card-actions>
          </v-form>
        </v-card>
      </v-dialog>



    </v-toolbar>

    <v-container fluid>
      <v-data-table
        :headers="headers"
        :items="desserts"
        :search="search"
        :loading="loading"
        :items-per-page="-1"
        class="elevation-1"
      >
        <template v-slot:body="{ items }">
          <tbody>
            <tr v-for="item in items" :key="item.file + item.section + item.key">
              <td>{{ item.file }}</td>
              <td>{{ item.section }}</td>
              <td>{{ item.key }}</td>
              <td>{{ item.type }}</td>
              <td>{{ item.value }}</td>
              <td>
                <v-icon small class="mr-2" @click="editItem(item)">mdi-pencil</v-icon>
                <v-icon small @click="viewItem(item)">mdi-eye</v-icon>
              </td>
            </tr>
          </tbody>
        </template>
      </v-data-table>
    </v-container>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  data: () => ({
    dialog: {
      item:false,
      domain:false
    },
    search: '',
    loading: true,
    clusters: [],
    currentCluster: '',
    headers: [
      { text: 'File', value: 'file', sortable: false },
      { text: 'Section', value: 'section', sortable: false },
      { text: 'Key', value: 'key', sortable: false },
      { text: 'Type', value: 'type', sortable: false },
      { text: 'Value', value: 'value', sortable: false },
      { text: 'Actions', value: 'actions', sortable: false }
    ],
    files: [],
    desserts: [],
    editable: true,
    editedIndex: -1,
    editedItem: {},
    newDomain: {},
    valid: true,
    emptyRules: [
      v => !!v || 'Field is required'
    ],
    keyTypes: [
      'bool',
      'int',
      'int64',
      'float64',
      'string',
      '[]string',
      '[]float64',
      '[]int64',
      '[]int',
      '[]bool',
      'map[string]string',
      'map[string]int',
      'map[string]struct{}',
      'map[int]string',
      'map[int]int',
      'map[int]struct{}',
      'map[string][]string',
      'json'
    ]
  }),

  computed: {
    itemDialog(){
      return this.dialog.item
    },
    formTitle() {
      return this.editedIndex === -1 ? 'New Item' : 'Edit Item'
    }
  },

  watch: {
    itemDialog(val) {
      val || this.close()
      if (val && this.editable && this.editedIndex === -1) {
        this.editedItem = {
          section: 'DEFAULT'
        }
      }
    }
  },

  created() {
    this.initialize()
  },

  methods: {
    initialize() {
      axios.get('/keeper/domains').then(res => {
        this.clusters = []
        res.data.data.forEach(item => {
          this.clusters.push(item.domain)
        })
        if (this.clusters.length > 0) {
          var cluster = this.clusters[0]
          if (localStorage.cluster) {
            cluster = localStorage.cluster
          }
          this.getConfig(cluster)
        }
      })
    },


    getConfig(cluster) {
      this.loading = true
      this.desserts = []
      axios.get('/config/' + cluster).then(res => {
        this.files = []
        res.data.data.forEach(file => {
          this.files.push(file.name)
          file.sections.forEach(section => {
            Object.keys(section.keys).forEach(name => {
              this.desserts.push({
                file: file.name,
                section: section.name,
                key: section.keys[name].key,
                type: section.keys[name].type,
                value: section.keys[name].raw_value
              })
            })
          })
        })
        this.currentCluster = cluster
        localStorage.cluster = cluster
        this.loading = false
      })
    },

    editItem(item) {
      this.editedIndex = this.desserts.indexOf(item)
      this.editedItem = Object.assign({}, item)
      this.dialog.item = true
    },

    viewItem(item) {
      this.editedIndex = this.desserts.indexOf(item)
      this.editedItem = Object.assign({}, item)
      this.dialog.item = true
      this.editable = false
    },

    close() {
      this.editable = true
      this.dialog.domain = false
      this.dialog.item = false
      this.editedIndex = -1
      this.$refs.form.reset()
    },

    saveDomain() {
      if (!this.$refs.domainForm.validate()) {
        return
      }
      let formData = new FormData();
      for (let file of this.newDomain.files) {
        formData.append("files", file, file.name)
      }
      axios.put('/keeper/'+this.newDomain.keeper+'/'+this.newDomain.name, formData).then(res => {
        if (res.data.code !== 0) {
          console.log('add new domain failed', res.data)
          return
        }
        this.initialize()
        this.close('domain')
      })
    },

    save() {
      if (!this.$refs.form.validate()) {
        return
      }
      this.editedItem.opcode = this.editedIndex === -1 ? 'add' : 'update'
      this.editedItem.domain = this.currentCluster
      axios.post('/config/' + this.currentCluster, this.editedItem).then(res => {
        if (res.data.code !== 0) {
          console.log('update config failed', res.data)
          return
        }
        this.initialize()
        this.close()
      })
    }
  }
}
</script>
